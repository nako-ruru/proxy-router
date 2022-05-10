package main

import (
	"errors"
	"fmt"
	"github.com/dop251/goja"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var statusMap = sync.Map{}

func check(config *Config)  {
	//make HealthTimeout string because tools that generates struct from yaml make it string
	//I do not want to replace its type with time.Duration everytime
	_, err := time.ParseDuration(config.Health.HealthTimeout)
	if err != nil {
		panic(err)
	}
	_, err = time.ParseDuration(config.Health.HealthInterval)
	if err != nil {
		panic(err)
	}

	var limit = make(chan Target, config.Health.Threads)

	go func() {
		for _, target := range config.Backends.Targets {
			limit <- target
		}
	}()
	for _, targetDoc := range config.Backends.TargetsFromDocument {
		go func() {
			targets, err := loadFromDocument(targetDoc)
			if err != nil {
				log.Printf("%+v", err)
				return
			}
			if len(targets) > 0 {
				go func() {
					for _, target := range targets {
						limit <- target
					}
				}()
			}
		}()
	}

	go func() {
		for {
			backend := <-limit
			log.Printf("checking %s", backend.Server)
			go func() {
				defer func() {
					timeout, _ := time.ParseDuration(config.Health.HealthInterval)
					time.Sleep(timeout)
					limit <- backend
				}()
				checkProxyEnd(backend, config.Health)
			}()
		}
	}()
}

func checkProxyEnd(backend Target, health Health) {
	start := time.Now()
	err := with_proxy(backend, health)
	if err != nil {
		log.Printf("%s: %+v", backend.Server, err)
		statusMap.Delete(backend)
	} else {
		elapsed := time.Since(start)
		log.Printf("%s health check ok within %d ms", backend.Server, int64(elapsed / time.Millisecond))
		statusMap.Store(backend, elapsed)
	}
}

func with_proxy(backend Target, health Health) error {
	var proxy Proxy
	if backend.Type == "http" {
		proxy = &HttpProxy{
			Url: "http://" + backend.Server,
		}
	}
	if proxy == nil {
		return errors.New("unsupported proxy type " + backend.Type)
	}
	client, err := proxy.NewClient()
	if err != nil {
		return err
	}
	timeout, _ := time.ParseDuration(health.HealthTimeout)
	client.Timeout = timeout
	req, err := http.NewRequest("GET", health.HealthURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", health.UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	if !responseCodeMath(health.HealthResponseStatus, resp.StatusCode) {
		return fmt.Errorf(`http status not match, "%s" expected, actual "%s"`, health.HealthResponseStatus, resp.Status)
	}
	return nil
}

func loadFromDocument(targetDoc TargetsFromDocument) ([]Target, error) {
	if targetDoc.ExtractType != "script" && targetDoc.ExtractType != "delimiter" {
		return nil, nil
	}
	var bytes []byte
	if targetDoc.FilePath != "" {
		var err error
		bytes, err = ioutil.ReadFile(targetDoc.FilePath)
		if err != nil {
			return nil, err
		}
	} else if targetDoc.URL != "" {
		resp, err := http.Get(targetDoc.URL)
		if err != nil {
			return nil, err
		}
		defer func(reader io.ReadCloser) {
			_ = reader.Close()
		}(resp.Body)
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf(`unexpected status code %d for %s`, resp.StatusCode, targetDoc.URL)
		}
		bytes, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf(`either file-path or url should be setL`)
	}
	var docContent = strings.TrimSpace(string(bytes))
	if docContent == "" {
		var path string
		if targetDoc.FilePath != "" {
			path = targetDoc.FilePath
		} else {
			path = targetDoc.URL
		}
		return nil, fmt.Errorf(`empty document of "%s"`, path)
	}
	if targetDoc.ExtractType == "script" {
		vm := goja.New()
		vm.SetFieldNameMapper(goja.TagFieldNameMapper("yaml", true))
		_, err := vm.RunString(targetDoc.ExtractScript)
		if err != nil {
			return nil, err
		}
		var extract func(string) []Target
		err = vm.ExportTo(vm.Get(targetDoc.ScriptEntrance), &extract)
		if err != nil {
			return nil, err
		}
		targets := extract(docContent)
		return targets, nil
	} else {
		elements := strings.Split(docContent, targetDoc.Delimiter)
		var targets []Target
		for _, element := range elements {
			targets = append(targets, Target{Server: element, Type: targetDoc.Type})
		}
		return targets, nil
	}
}

func responseCodeMath(pattern string, code int) bool {
	codeText := strconv.Itoa(code)
	if len(pattern) != len(codeText) {
		return false
	}
	for i := 0; i < len(codeText); i++ {
		if pattern[i] != codeText[i] && pattern[i] != 'x' {
			return false
		}
	}
	return true
}