package main

import (
	"encoding/base64"
	"fmt"
	"github.com/xzycn/goproxy"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var selector = &MinimumResponseDuration{}

const (
	ProxyAuthHeader = "Proxy-Authorization"
)

func SetBasicAuth(username, password string, req *http.Request) {
	if true {
		return
	}
	req.Header.Set(ProxyAuthHeader, fmt.Sprintf("Basic %s", basicAuth(username, password)))
}

func basicAuth(username, password string) string {
	return base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}


func GetBasicAuth(req *http.Request) (username, password string, ok bool) {
	auth := req.Header.Get(ProxyAuthHeader)
	if auth == "" {
		return
	}

	const prefix = "Basic "
	if !strings.HasPrefix(auth, prefix) {
		return
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return
	}
	return cs[:s], cs[s+1:], true
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
	username, password := "foo", "bar"
	middleProxy.OnRequest().HandleConnectFunc(func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		c := selector.Select()
		if c == nil {
			return goproxy.RejectConnect, host
		}
		connectReqHandler := func(req *http.Request) {
			SetBasicAuth(username, password, req)
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
		err := http.ListenAndServe(formatAddress(proxyEnd.Address), middleProxy)
		if err != nil {
			log.Printf("serving middle proxy server fails in listening %s", proxyEnd.Address)
		}
	}()
}

func formatAddress(address string) string {
	match, _ := regexp.MatchString(`^\d+$`, address)
	if match {
		return ":" + address
	}
	return address
}

func parseBool(flag string) bool {
	regex := regexp.MustCompile(`^(y|t|yes|true|1)$`)
	return regex.MatchString(flag)
}

func NewConnectDialToProxyWithHandler(middle *goproxy.ProxyHttpServer, connectReqHandler func(req *http.Request)) func(network, addr string) (net.Conn, error) {
	c := selector.Select()
	if c == nil {
		return nil
	}
	endproxy := c.Server
	return middle.NewConnectDialToProxyWithHandler(endproxy, connectReqHandler)
}