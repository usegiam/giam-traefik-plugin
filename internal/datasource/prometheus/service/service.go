package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/usegiam/giam-traefik-plugin/internal/datasource/prometheus"
	"io"
	"net/http"

	"github.com/usegiam/giam-traefik-plugin/pkg/log"
)

type service struct {
	apiUrl string
	apiKey string
	logger *log.Logger
}

type Deps struct {
	APIUrl string
	APIKey string
	Logger *log.Logger
}

func New(deps *Deps) prometheus.Service {
	return &service{apiUrl: deps.APIUrl, apiKey: deps.APIKey, logger: deps.Logger}
}

func (s *service) AuthorizeQuery(payload *prometheus.AuthorizeQueryReq) (*prometheus.AuthorizedQueryResp, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, s.apiUrl+"/prometheus/query/authorize", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", s.apiKey)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	s.logger.Debugf(
		"giam prometheus authorize query resp status code: %v, resp body: %s",
		resp.StatusCode,
		string(respBody),
	)

	var queryResp prometheus.AuthorizedQueryResp

	err = json.Unmarshal(respBody, &queryResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode giam prometheus authorize query resp: %w", err)
	}

	queryResp.StatusCode = resp.StatusCode

	return &queryResp, nil
}

func (s *service) FilterSeries(payload *prometheus.FilterSeriesReq) (*prometheus.FilterSeriesResp, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, s.apiUrl+"/prometheus/series/filter", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", s.apiKey)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	s.logger.Debugf(
		"giam prometheus filter series resp status code: %v, resp body: %s",
		resp.StatusCode,
		string(respBody),
	)

	var filterSeriesResp prometheus.FilterSeriesResp

	err = json.Unmarshal(respBody, &filterSeriesResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode giam prometheus filter series resp: %w", err)
	}

	filterSeriesResp.StatusCode = resp.StatusCode

	return &filterSeriesResp, nil
}
