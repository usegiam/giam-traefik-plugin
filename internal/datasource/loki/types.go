package loki

import (
	"github.com/usegiam/giam-traefik-plugin/pkg/grafana"
)

type Service interface {
	AuthorizeQuery(payload *AuthorizeQueryReq) (*AuthorizedQueryResp, error)
	FilterLabelValues(payload *FilterLabelValuesReq) (*FilterLabelValuesResp, error)
	FilterSeries(payload *FilterSeriesReq) (*FilterSeriesResp, error)
}

type AuthorizedQueryResp struct {
	Queries    []interface{} `json:"queries"`
	Message    string        `json:"message"`
	StatusCode int           `json:"status_code"`
}

type AuthorizeQueryReq struct {
	User    *grafana.User   `json:"user"`
	Teams   []*grafana.Team `json:"teams"`
	Queries []interface{}   `json:"queries"`
}

type FilterSeriesReq struct {
	User       *grafana.User       `json:"user"`
	Teams      []*grafana.Team     `json:"teams"`
	Series     []map[string]string `json:"series"`
	Datasource grafana.Datasource  `json:"datasource"`
}

type FilterSeriesResp struct {
	Data       []map[string]string `json:"data"`
	StatusCode int                 `json:"status_code"`
}

type Label struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

type FilterLabelValuesReq struct {
	User       *grafana.User      `json:"user"`
	Teams      []*grafana.Team    `json:"teams"`
	Label      *Label             `json:"label"`
	Datasource grafana.Datasource `json:"datasource"`
}

type FilterLabelValuesResp struct {
	Data       []string `json:"data"`
	StatusCode int      `json:"status_code"`
}
