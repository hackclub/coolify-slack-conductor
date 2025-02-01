package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
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

func main() {
	fmt.Println("Starting")
	loadDestinations()

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
		destUrl, _ := url.Parse(MainDestination.url)
		log.Println("Sending to main:", destUrl.String())
		req.SetURL(destUrl)
	}

	proxy := &httputil.ReverseProxy{Rewrite: rewrite}
	proxy.Transport = DebugTransport{}

	log.Fatalln(http.ListenAndServe(":8080", proxy))
}
