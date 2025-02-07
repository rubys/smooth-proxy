package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

const (
	targetHost = "https://smooth.fly.dev"
	proxyPort  = ":8080"
)

func main() {
	// Parse the target URL
	target, err := url.Parse(targetHost)
	if err != nil {
		log.Fatal(err)
	}

	// Create a reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Store the default director
	defaultDirector := proxy.Director

	// Create a custom director that will modify the request
	proxy.Director = func(req *http.Request) {
		defaultDirector(req)

		// Get all Authorization headers
		authHeaders := req.Header.Values("Authorization")
		if len(authHeaders) > 1 {
			// Remove all Authorization headers and set only the first one
			req.Header.Del("Authorization")
			req.Header.Set("Authorization", authHeaders[0])
		}

		// Set the host to match the target
		req.Host = target.Host
	}

	// Create a handler that logs requests
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Proxying request: %s %s", r.Method, r.URL.Path)
		proxy.ServeHTTP(w, r)
	})

	// Start the server
	log.Printf("Starting proxy server on %s, forwarding to %s", proxyPort, targetHost)
	if err := http.ListenAndServe(proxyPort, handler); err != nil {
		log.Fatal(err)
	}
}
