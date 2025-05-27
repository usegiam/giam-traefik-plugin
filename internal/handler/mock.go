package handler

import (
	"net/http"
)

type mockHandler struct {
	match      bool
	shouldStop bool
	called     bool
	modifyReq  func(*http.Request)
	// modifyResp is only used when shouldStop is true.
	modifyResp func(http.ResponseWriter)
}

func (m *mockHandler) Match(req *http.Request) bool {
	return m.match
}

func (m *mockHandler) Handle(rw http.ResponseWriter, req *http.Request, next http.Handler) {
	m.called = true

	if m.modifyReq != nil {
		m.modifyReq(req)
	}

	if m.shouldStop {
		if m.modifyResp != nil {
			m.modifyResp(rw)
		}
		return
	}

	next.ServeHTTP(rw, req)
}
