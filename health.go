package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var statusMap = sync.Map{}

type ProxyEnd struct {
	OnDown func(*ProxyEnd)
}

type HealthManager struct {
	ProxyEnds []*ProxyEnd
	AvailableProxyEnds []*ProxyEnd
	Current *ProxyEnd
	OnDown func(*ProxyEnd)
	OnUp func(*ProxyEnd)
}

func check(config *Config)  {
	var limit = make(chan BackendC, config.Health.Threads)

	go func() {
		for _, backend := range config.Backends {
			limit <- backend
		}
	}()

	go func() {
		for {
			backend := <-limit
			log.Printf("%s", (backend).Server)
			go func() {
				defer func() {
					time.Sleep(config.Health.HealthInterval)
					limit <- backend
				}()
				checkProxyEnd(backend, config)
			}()
		}
	}()
}

func checkProxyEnd(backend BackendC, config *Config) {
	start := time.Now()
	err := with_proxy(backend, &config.Health)
	if err != nil {
		log.Printf("%s: %+v", backend.Server, err)
		statusMap.Delete(&backend)
	} else {
		elapsed := time.Since(start)
		log.Printf("%s health check ok", backend.Server)
		statusMap.Store(&backend, elapsed)
	}
}

func with_proxy(backend BackendC, health *HealthC) error {
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
	client.Timeout = health.HealthTimeout
	req, err := http.NewRequest("GET", health.HealthURI, nil)
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