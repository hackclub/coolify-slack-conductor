package main

import (
	"log"
	"net/http"
	"net/url"
)

func setRequestPath(req *http.Request, target *url.URL) {
	// Set URL. Inspired by ProxyRequest.SetURL
	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
	// ProxyRequest.SetURL merges Path, RawPath, and queries, but i don't
	req.URL.Path = target.Path
	req.URL.RawPath = target.RawPath
	req.URL.RawQuery = target.RawQuery
	req.Host = ""
}

func mirrorRequest(req http.Request, destUrl string) {
	target, _ := url.Parse(destUrl)
	setRequestPath(&req, target)

	req.RequestURI = "" // Can not be set for client requests

	_, err := http.DefaultClient.Do(&req)
	if err != nil {
		log.Println(err)
	}
}
