package main

import (
	"log"
	"net/http"
	"net/url"
)

func setRequestPath(req *http.Request, target *url.URL) {
	req.URL = target
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
