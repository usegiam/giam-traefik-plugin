package service

import (
	"github.com/usegiam/giam-traefik-plugin/internal/datasource/loki"
)

type Mock struct {
	Error                 error
	AuthorizedQueryResp   *loki.AuthorizedQueryResp
	FilterSeriesResp      *loki.FilterSeriesResp
	FilterLabelValuesResp *loki.FilterLabelValuesResp
}

func (m *Mock) AuthorizeQuery(payload *loki.AuthorizeQueryReq) (*loki.AuthorizedQueryResp, error) {
	return m.AuthorizedQueryResp, m.Error
}

func (m *Mock) FilterSeries(payload *loki.FilterSeriesReq) (*loki.FilterSeriesResp, error) {
	return m.FilterSeriesResp, m.Error
}

func (m *Mock) FilterLabelValues(payload *loki.FilterLabelValuesReq) (*loki.FilterLabelValuesResp, error) {
	return m.FilterLabelValuesResp, m.Error
}
