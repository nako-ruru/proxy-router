package main

import (
	"net/http"
	"net/url"
)

type Proxy interface {
	NewClient() (*http.Client, error)
}

type HttpProxy struct {
	Url string
}

func (_this *HttpProxy) NewClient() (*http.Client, error) {
	//creating the proxyURL
	proxyURL, err := url.Parse(_this.Url)
	if err != nil {
		return nil, err
	}
	//adding the proxy settings to the Transport object
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
	//adding the Transport object to the http Client
	return &http.Client{
		Transport: transport,
	}, nil
}