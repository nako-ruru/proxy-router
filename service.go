package main

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"github.com/elazarl/goproxy"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
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
	middleProxy.ConnectDialWithReq = func(req *http.Request, network string, addr string) (net.Conn, error) {
		c := selector.Select()
		if c == nil {
			return nil, fmt.Errorf("no Available communication circuit")
		}
		connectReqHandler := func(req *http.Request) {
			SetBasicAuth(c.Username, c.Password, req)
		}
		handler := middleProxy.NewConnectDialToProxyWithHandler("http://"+c.Server, connectReqHandler)
		return handler(network, addr)
	}
	middleProxy.Verbose = parseBool(config.Server.Verbose)

	middleProxy.Tr.Proxy = func(req *http.Request) (*url.URL, error) {
		c := selector.Select()
		if c == nil {
			return nil, fmt.Errorf("no Available communication circuit")
		}
		return url.Parse("http://" + c.Server)
	}
	for _, proxyEnd := range config.Server.Listen {
		listen(proxyEnd, middleProxy)
	}

	if config.Server.Prof != "" {
		go func() {
			log.Println(http.ListenAndServe(formatAddress(config.Server.Prof), nil))
		}()
	}
}

func listen(proxyEnd Listen, middleProxy *goproxy.ProxyHttpServer) {
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
