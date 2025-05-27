package handler

import (
	"net/http"
)

type Handler interface {
	Match(req *http.Request) bool
	Handle(rw http.ResponseWriter, req *http.Request, next http.Handler)
}
