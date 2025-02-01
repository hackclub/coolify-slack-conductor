package main

import (
	"log"
	"net/http"
	"net/url"
)

func setRequestPath(req *http.Request, target *url.URL) {
	// Set URL. Inspired by ProxyRequest.SetURL
	targetQuery := target.RawQuery
	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
	// ProxyRequest.SetURL joins Path and RawPath, but i don't
	req.URL.Path = target.Path
	req.URL.RawPath = target.RawPath
	if targetQuery == "" || req.URL.RawQuery == "" {
		req.URL.RawQuery = targetQuery + req.URL.RawQuery
	} else {
		req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
	}
	req.Host = ""
}

func mirrorRequest(req http.Request, destUrl string) {
	target, _ := url.Parse(destUrl)
	setRequestPath(&req, target)

	req.RequestURI = "" // Can not be set for client requests

	_, err := http.DefaultClient.Do(&req)
	if err != nil {
		log.Fatalln(err)
	}
}
