package service

import "github.com/usegiam/giam-traefik-plugin/internal/authorization"

type Mock struct {
	Error                   error
	AuthorizeDatasourceResp *authorization.AuthorizeDatasourceResp
}

func (m *Mock) AuthorizeQuery(_ *authorization.AuthorizeDatasourceReq) (*authorization.AuthorizeDatasourceResp, error) {
	return m.AuthorizeDatasourceResp, m.Error
}
