package main

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"
)

func TestProxyHeaderDeduplication(t *testing.T) {
	// Create a test server that will act as our target
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get all Authorization headers from the request
		authHeaders := r.Header.Values("Authorization")
		
		// Verify we only have one Authorization header
		if len(authHeaders) != 1 {
			t.Errorf("Expected exactly 1 Authorization header, got %d", len(authHeaders))
		}

		// Verify the value is what we expect
		if len(authHeaders) > 0 && authHeaders[0] != "Bearer token1" {
			t.Errorf("Expected Authorization header to be 'Bearer token1', got '%s'", authHeaders[0])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	// Create a test request with duplicate Authorization headers
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Add("Authorization", "Bearer token1")
	req.Header.Add("Authorization", "Bearer token2")

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Parse the test server URL
	target, err := url.Parse(testServer.URL)
	if err != nil {
		t.Fatal(err)
	}

	// Create the proxy pointing to our test server
	proxy := httputil.NewSingleHostReverseProxy(target)
	defaultDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		defaultDirector(req)
		
		// Get all Authorization headers
		authHeaders := req.Header.Values("Authorization")
		if len(authHeaders) > 1 {
			req.Header.Del("Authorization")
			req.Header.Set("Authorization", authHeaders[0])
		}
		req.Host = target.Host
	}

	// Send the request through our proxy
	proxy.ServeHTTP(rr, req)

	// Check the response status
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
