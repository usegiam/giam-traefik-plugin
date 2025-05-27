package prometheus

import (
	"github.com/usegiam/giam-traefik-plugin/pkg/grafana"
)

type Service interface {
	AuthorizeQuery(payload *AuthorizeQueryReq) (*AuthorizedQueryResp, error)
	FilterSeries(payload *FilterSeriesReq) (*FilterSeriesResp, error)
}

type Repo interface {
	GetMetrics(session string, uid string) ([]*Metric, error)
}

type Metric struct {
	Name string `json:"name"`
}

type AuthorizedQueryResp struct {
	Queries    []interface{} `json:"queries"`
	Message    string        `json:"message"`
	StatusCode int           `json:"status_code"`
}

type AuthorizeQueryReq struct {
	User    interface{}     `json:"user"`
	Teams   []*grafana.Team `json:"teams"`
	Queries []interface{}   `json:"queries"`
}

type FilterSeriesReq struct {
	User       interface{}         `json:"user"`
	Teams      []*grafana.Team     `json:"teams"`
	Series     []map[string]string `json:"series"`
	Datasource grafana.Datasource  `json:"datasource"`
}

type FilterSeriesResp struct {
	Data       []map[string]string `json:"data"`
	StatusCode int                 `json:"status_code"`
}
