package main

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"github.com/xzycn/goproxy"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

var selector = &MinimumResponseDuration{}

const (
	ProxyAuthHeader = "Proxy-Authorization"
)

func SetBasicAuth(username, password string, req *http.Request) {
	if username != "" {
		req.Header.Set(ProxyAuthHeader, fmt.Sprintf("Basic %s", basicAuth(username, password)))
	}
}

func basicAuth(username, password string) string {
	return base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}

func start(config *Config) {
	// start middle proxy server
	middleProxy := goproxy.NewProxyHttpServer()
	middleProxy.ConnectDial = nil
	middleProxy.Verbose = parseBool(config.Server.Verbose)

	middleProxy.Tr.Proxy = func(req *http.Request) (*url.URL, error) {
		c := selector.Select()
		if c == nil {
			panic("No Available communication circuit!")
		}
		return url.Parse("http://" + c.Server)
	}
	middleProxy.OnRequest().HandleConnectFunc(func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		c := selector.Select()
		if c == nil {
			return goproxy.RejectConnect, host
		}
		connectReqHandler := func(req *http.Request) {
			SetBasicAuth( c.Username, c.Password, req)
		}
		ctx.ConnectDial = middleProxy.NewConnectDialToProxyWithHandler("http://"+c.Server, connectReqHandler)
		return goproxy.OkConnect, host
	})
	for _, proxyEnd := range config.Server.Listen {
		listen(proxyEnd, middleProxy)
	}

	time.Sleep(1 * time.Second)
}

func listen(proxyEnd Listen, middleProxy *goproxy.ProxyHttpServer)  {
	go func() {
		log.Printf("serving middle proxy server start listening %s", proxyEnd.Address)
		if parseBool(proxyEnd.AutoCert) {
			mux := http.NewServeMux()
			mux.Handle("/", middleProxy)
			certManager := autocert.Manager{
				Prompt: autocert.AcceptTOS,
				Cache:  autocert.DirCache("certs"),
			}
			server := &http.Server{
				Addr:    formatAddress(proxyEnd.Address),
				Handler: mux,
				TLSConfig: &tls.Config{
					GetCertificate: certManager.GetCertificate,
				},
			}
			err := server.ListenAndServeTLS("", "")
			if err != nil {
				log.Printf("serving middle proxy server fails in listening %s", proxyEnd.Address)
			}
		} else if proxyEnd.CertFile != "" && proxyEnd.KeyFile != "" {
			err := http.ListenAndServeTLS(formatAddress(proxyEnd.Address), proxyEnd.CertFile, proxyEnd.KeyFile, middleProxy)
			if err != nil {
				log.Printf("serving middle proxy server fails in listening %s", proxyEnd.Address)
			}
		} else {
			err := http.ListenAndServe(formatAddress(proxyEnd.Address), middleProxy)
			if err != nil {
				log.Printf("serving middle proxy server fails in listening %s", proxyEnd.Address)
			}
		}
	}()
}

func formatAddress(address string) string {
	matched, _ := regexp.MatchString(`^\d+$`, address)
	if matched {
		return ":" + address
	}
	return address
}

func parseBool(flag string) bool {
	matched, _ := regexp.MatchString(`^(y|t|yes|true|1)$`, flag)
	return matched
}