package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/usegiam/giam-traefik-plugin/pkg/testutil/assert"
)

func TestChainHandlers_NoStop(t *testing.T) {
	finalCalled := false
	finalHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		finalCalled = true
		response := map[string]string{"message": "final handler executed"}
		json.NewEncoder(rw).Encode(response)
	})

	handler1 := &mockHandler{
		match:      true,
		shouldStop: false,
		modifyReq: func(req *http.Request) {
			req.Header.Set("X-Test-Header", "Modified by Handler1")
		},
	}

	handler2 := &mockHandler{
		match:      true,
		shouldStop: false,
	}

	chained := ChainHandlers(finalHandler, handler1, handler2)

	req := httptest.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte(`{"original": "request"}`)))
	res := httptest.NewRecorder()
	chained.ServeHTTP(res, req)

	assert.True(t, handler1.called)
	assert.True(t, handler2.called)
	assert.True(t, finalCalled)

	var responseBody map[string]string
	err := json.NewDecoder(res.Body).Decode(&responseBody)
	assert.NoError(t, err)
	assert.Equal(t, "final handler executed", responseBody["message"])
}

func TestChainHandlers_WithStop(t *testing.T) {
	finalCalled := false
	finalHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		finalCalled = true
		response := map[string]string{"message": "final handler executed"}
		json.NewEncoder(rw).Encode(response)
	})

	handler1 := &mockHandler{
		match:      true,
		shouldStop: false,
	}

	handlerStop := &mockHandler{
		match:      true,
		shouldStop: true,
		modifyResp: func(rw http.ResponseWriter) {
			response := map[string]string{"message": "stopped by handlerStop"}
			json.NewEncoder(rw).Encode(response)
		},
	}

	handlerAfter := &mockHandler{
		match:      true,
		shouldStop: false,
	}

	chained := ChainHandlers(finalHandler, handler1, handlerStop, handlerAfter)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	chained.ServeHTTP(res, req)

	assert.True(t, handler1.called)
	assert.True(t, handlerStop.called)
	assert.False(t, handlerAfter.called)
	assert.False(t, finalCalled)

	var responseBody map[string]string
	err := json.NewDecoder(res.Body).Decode(&responseBody)
	assert.NoError(t, err)
	assert.Equal(t, "stopped by handlerStop", responseBody["message"])
}
