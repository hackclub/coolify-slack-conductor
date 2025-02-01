package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
)

type DebugTransport struct{}

func (DebugTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	b, err := httputil.DumpRequestOut(r, false)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(b))
	return http.DefaultTransport.RoundTrip(r)
}

type Destination struct {
	url    string
	regexp string
}

func (dest Destination) Matches(body string) bool {
	match, _ := regexp.MatchString(dest.regexp, body)
	return match
}

func destinations(body string) []Destination {
	config := []Destination{
		{
			url:    "https://webhook.site/a32c93b6-747e-4f29-8b2c-df6eb59b6fb2?afm",
			regexp: "\\*\\*Project:\\*\\* gary@mfa\\\\n",
		},
		{
			url:    "https://webhook.site/a32c93b6-747e-4f29-8b2c-df6eb59b6fb2?danlog",
			regexp: "goland",
		},
	}

	var results []Destination
	for _, dest := range config {
		if dest.Matches(body) {
			results = append(results, dest)
		}
	}
	return results
}

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

func main() {
	fmt.Println("Starting")
	rewrite := func(req *httputil.ProxyRequest) {
		// Read body
		body, _ := io.ReadAll(req.In.Body)
		log.Println(string(body))

		req.SetXForwarded() // set headers on Out request

		dests := destinations(string(body))

		// Reset body back to original
		req.In.Body = io.NopCloser(bytes.NewBuffer(body))
		req.Out.Body = io.NopCloser(bytes.NewBuffer(body))

		// Send to relevant destinations
		for _, dest := range dests {
			log.Println("Mirroring to", dest.url)

			// Cloned request for the mirror
			mReq := req.Out.Clone(req.Out.Context())
			mReq.Body = io.NopCloser(bytes.NewReader(body))

			mirrorRequest(*mReq, dest.url)
		}

		// Always send to main destination
		destUrl, _ := url.Parse("https://webhook.site/a32c93b6-747e-4f29-8b2c-df6eb59b6fb2")
		log.Println("Sending to main:", destUrl.String())
		req.SetURL(destUrl)
	}

	proxy := &httputil.ReverseProxy{Rewrite: rewrite}
	proxy.Transport = DebugTransport{}

	log.Fatalln(http.ListenAndServe(":80", proxy))
}
