package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
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

	authKey := os.Getenv("AUTH_KEY")
	if authKey == "" {
		log.Fatalln("Missing AUTH_KEY environment variable")
	}
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "GET" && req.RequestURI == "/" {
			http.Redirect(w, req, "https://github.com/hackclub/coolify-slack-conductor", 302)
			return
		}

		// Having authentication here prevents ppl from spamming our slack channels
		if len(req.URL.Query()["key"]) == 0 || req.URL.Query()["key"][0] != authKey {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			log.Println("Invalid authentication:", req.RequestURI)
			return
		}

		proxy.ServeHTTP(w, req)
	})

	log.Fatalln(http.ListenAndServe(":8080", nil))
}
