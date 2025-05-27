package authorization

import "github.com/usegiam/giam-traefik-plugin/pkg/grafana"

type Service interface {
	AuthorizeQuery(payload *AuthorizeDatasourceReq) (*AuthorizeDatasourceResp, error)
}

type AuthorizeDatasourceReq struct {
	User    *grafana.User   `json:"user"`
	Teams   []*grafana.Team `json:"teams"`
	Queries []interface{}   `json:"queries"`
}

type AuthorizeDatasourceResp struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}
