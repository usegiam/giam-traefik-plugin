package types

import (
	"bytes"
	"net/http"
)

type ResponseWriter struct {
	http.ResponseWriter
	Body   *bytes.Buffer
	Status int
}

func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.Status = statusCode
}

func (rw *ResponseWriter) Write(b []byte) (int, error) {
	return rw.Body.Write(b)
}
