package handler_test

import (
	"assignment_02/handler"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStatusHandler(t *testing.T) {
	// Test serverHandler
	ts := httptest.NewServer(http.HandlerFunc(handler.HandleStatus))
	defer ts.Close()

	t.Log("Server URL: ", ts.URL)

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nBody: %s", err, body)
	}

	if result["version"] != "v1" {
		t.Errorf("Expected version 'v1', got '%v'", result["version"])
	}
}
