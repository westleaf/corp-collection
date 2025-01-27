package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServe(t *testing.T) {
	dirname := "."

	// Create a new server
	server := httptest.NewServer(http.FileServer(http.Dir(dirname)))
	defer server.Close()

	// Make a request to the server
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("failed to make GET request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %v", resp.Status)
	}
}
