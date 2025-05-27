package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/usegiam/giam-traefik-plugin/internal/datasource/loki"
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

func New(deps *Deps) loki.Service {
	return &service{apiUrl: deps.APIUrl, apiKey: deps.APIKey, logger: deps.Logger}
}

func (s *service) AuthorizeQuery(payload *loki.AuthorizeQueryReq) (*loki.AuthorizedQueryResp, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, s.apiUrl+"/loki/query/authorize", bytes.NewBuffer(payloadBytes))
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
		"giam loki authorized query resp status code: %v, resp body: %s",
		resp.StatusCode,
		string(respBody),
	)

	var queryResp loki.AuthorizedQueryResp

	err = json.Unmarshal(respBody, &queryResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode giam loki authorize query resp: %w", err)
	}

	queryResp.StatusCode = resp.StatusCode

	return &queryResp, nil
}

func (s *service) FilterSeries(payload *loki.FilterSeriesReq) (*loki.FilterSeriesResp, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, s.apiUrl+"/loki/series/filter", bytes.NewBuffer(payloadBytes))
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

	s.logger.Debugf("giam loki filter series resp status code: %v, resp body: %s", resp.StatusCode, string(respBody))

	var filterSeriesResp loki.FilterSeriesResp

	err = json.Unmarshal(respBody, &filterSeriesResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode giam loki filter series resp: %w", err)
	}

	filterSeriesResp.StatusCode = resp.StatusCode

	return &filterSeriesResp, nil
}

func (s *service) FilterLabelValues(payload *loki.FilterLabelValuesReq) (*loki.FilterLabelValuesResp, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, s.apiUrl+"/loki/label/filter", bytes.NewBuffer(payloadBytes))
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
		"giam loki filter label values resp status code: %v, resp body: %s",
		resp.StatusCode,
		string(respBody),
	)

	var filterLabelValuesResp loki.FilterLabelValuesResp

	err = json.Unmarshal(respBody, &filterLabelValuesResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode giam loki filter label values resp: %w", err)
	}

	filterLabelValuesResp.StatusCode = resp.StatusCode

	return &filterLabelValuesResp, nil
}
