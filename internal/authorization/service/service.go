package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/usegiam/giam-traefik-plugin/internal/authorization"
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

func NewService(deps *Deps) authorization.Service {
	return &service{apiUrl: deps.APIUrl, apiKey: deps.APIKey, logger: deps.Logger}
}

func (s *service) AuthorizeQuery(
	payload *authorization.AuthorizeDatasourceReq,
) (*authorization.AuthorizeDatasourceResp, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, s.apiUrl+"/authorization/datasource", bytes.NewBuffer(payloadBytes))
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
		"giam authorized datasource resp status code: %v, resp body: %s",
		resp.StatusCode,
		string(respBody),
	)

	var authorizationResp authorization.AuthorizeDatasourceResp

	err = json.Unmarshal(respBody, &authorizationResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode giam authorize datasource resp: %w", err)
	}

	authorizationResp.StatusCode = resp.StatusCode

	return &authorizationResp, nil
}
