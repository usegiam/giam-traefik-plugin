package handler

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"regexp"

	"github.com/usegiam/giam-traefik-plugin/internal/datasource/prometheus"
	"github.com/usegiam/giam-traefik-plugin/internal/handler"
	"github.com/usegiam/giam-traefik-plugin/internal/types"
	"github.com/usegiam/giam-traefik-plugin/pkg/grafana"
	"github.com/usegiam/giam-traefik-plugin/pkg/log"
)

const seriesEndpointPattern = `^/api/datasources/uid/([a-zA-Z0-9-]+)/resources/api/v1/series`

var seriesEndpointRegexExp = regexp.MustCompile(seriesEndpointPattern)

type SeriesHandler struct {
	grafanaRepo grafana.Repo
	logger      *log.Logger
	service     prometheus.Service
}

type SeriesHandlerDeps struct {
	GrafanaRepo grafana.Repo
	Logger      *log.Logger
	Service     prometheus.Service
}

func NewSeriesHandler(deps *SeriesHandlerDeps) handler.Handler {
	return &SeriesHandler{service: deps.Service, logger: deps.Logger}
}

func (l *SeriesHandler) Match(req *http.Request) bool {
	return seriesEndpointRegexExp.MatchString(req.RequestURI)
}

func (l *SeriesHandler) Handle(rw http.ResponseWriter, req *http.Request, next http.Handler) {
	l.logger.Debug("instantiated a prometheus series filter")

	w := &types.ResponseWriter{
		ResponseWriter: rw,
		Body:           &bytes.Buffer{},
		Status:         http.StatusOK,
	}

	regx := regexp.MustCompile(seriesEndpointPattern)
	matches := regx.FindStringSubmatch(req.RequestURI)

	grafanaSession, err := req.Cookie("grafana_session")
	if err != nil {
		http.Error(rw, "Forbidden", http.StatusForbidden)

		return
	}

	next.ServeHTTP(w, req)

	var reader io.ReadCloser

	if w.Header().Get("Content-Encoding") == "gzip" {
		reader, err = gzip.NewReader(w.Body)
		if err != nil {
			http.Error(rw, "Internal server error", http.StatusPreconditionFailed)

			return
		}

		defer reader.Close()
	} else {
		reader = io.NopCloser(w.Body)
	}

	decompressedBody, err := io.ReadAll(reader)
	if err != nil {
		http.Error(rw, "Internal server error", http.StatusPreconditionFailed)

		return
	}

	var grafanaResp grafana.SeriesReq

	err = json.Unmarshal(decompressedBody, &grafanaResp)
	if err != nil {
		http.Error(rw, "Internal server error", http.StatusPreconditionFailed)

		return
	}

	user, err := l.grafanaRepo.GetUser(grafanaSession.Value)
	if err != nil {
		http.Error(rw, "User doesn't exits", http.StatusBadRequest)

		return
	}

	teams, err := l.grafanaRepo.GetUserTeams(grafanaSession.Value, user.ID)
	if err != nil {
		l.logger.Debugf("user doesn't have any team, err: %v", err)

		http.Error(rw, "User not assigned to any team", http.StatusBadRequest)

		return
	}

	resp, err := l.service.FilterSeries(&prometheus.FilterSeriesReq{
		User:       user,
		Teams:      teams,
		Series:     grafanaResp.Series,
		Datasource: grafana.Datasource{UID: matches[1]}, // Uid exists in first index of the regx
	})
	if err != nil {
		l.logger.Debugf("unable to send prometheus filter series request to Giam, err: %v", err)

		http.Error(rw, "Unable to communicate with Giam service", http.StatusPreconditionFailed)

		return
	}

	grafanaResp.Series = resp.Data

	rw.Header().Set("Content-Encoding", "gzip")
	rw.Header().Set("Content-Type", "application/json")

	gz := gzip.NewWriter(rw)
	defer gz.Close()

	encoder := json.NewEncoder(gz)
	if err := encoder.Encode(grafanaResp); err != nil {
		http.Error(rw, "Failed to encode the response", http.StatusInternalServerError)

		return
	}
}
