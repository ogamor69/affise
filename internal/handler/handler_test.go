package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Test for method not allowed.
func TestHandlerMethodNotAllowed(t *testing.T) {
	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Handler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

// Test for invalid JSON payload.
func TestHandlerInvalidJSONPayload(t *testing.T) {
	req, err := http.NewRequest("POST", "", bytes.NewBuffer([]byte("invalid json")))
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Handler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// Test for URL limit exceeded.
func TestHandlerURLLimitExceeded(t *testing.T) {
	urls := make([]string, 21)
	for i := 0; i < 21; i++ {
		urls[i] = "http://example.com"
	}
	payload := requestPayload{URLs: urls}
	jsonPayload, err := json.Marshal(payload)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "", bytes.NewBuffer(jsonPayload))
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Handler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// Test for successful handler execution.
func TestHandlerSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.Write([]byte("mock response"))
	}))
	defer ts.Close()

	payload := requestPayload{URLs: []string{ts.URL}}
	jsonPayload, err := json.Marshal(payload)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "", bytes.NewBuffer(jsonPayload))
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Handler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp responsePayload
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)

	expectedData := map[string]string{ts.URL: "mock response"}
	assert.Equal(t, expectedData, resp.Data)
}
