package mocks

import (
	"bytes"
	"io"
	"net/http"
)

type NextHandler struct {
	Called       bool
	ReceivedBody []byte
	ReqBody      []byte
	RespBody     []byte
	HeaderMap    http.Header
}

func (m *NextHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	m.Called = true
	body, _ := io.ReadAll(req.Body)
	m.ReceivedBody = body

	if m.HeaderMap != nil {
		for key, values := range m.HeaderMap {
			for _, value := range values {
				rw.Header().Add(key, value)
			}
		}
	}

	if m.ReqBody != nil {
		req.Body = io.NopCloser(bytes.NewBuffer(m.ReqBody))
	} else {
		req.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	if m.RespBody != nil {
		rw.Write(m.RespBody)
	}
}
