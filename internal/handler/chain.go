package handler

import (
	"net/http"
)

func ChainHandlers(finalHandler http.Handler, handlers ...Handler) http.Handler {
	chained := finalHandler

	for i := len(handlers) - 1; i >= 0; i-- {
		h := handlers[i]

		chained = func(h Handler, next http.Handler) http.Handler {
			return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				if h.Match(req) {
					h.Handle(rw, req, next)
				} else {
					next.ServeHTTP(rw, req)
				}
			})
		}(h, chained)
	}

	return chained
}
