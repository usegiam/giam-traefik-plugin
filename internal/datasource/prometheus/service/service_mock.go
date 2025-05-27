package service

import "github.com/usegiam/giam-traefik-plugin/internal/datasource/prometheus"

type Mock struct {
	Error               error
	AuthorizedQueryResp *prometheus.AuthorizedQueryResp
	FilterSeriesResp    *prometheus.FilterSeriesResp
}

func (m *Mock) AuthorizeQuery(payload *prometheus.AuthorizeQueryReq) (*prometheus.AuthorizedQueryResp, error) {
	return m.AuthorizedQueryResp, m.Error
}

func (m *Mock) FilterSeries(payload *prometheus.FilterSeriesReq) (*prometheus.FilterSeriesResp, error) {
	return m.FilterSeriesResp, m.Error
}
