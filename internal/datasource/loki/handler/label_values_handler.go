package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"

	"github.com/usegiam/giam-traefik-plugin/internal/datasource"
	"github.com/usegiam/giam-traefik-plugin/internal/datasource/loki"
	"github.com/usegiam/giam-traefik-plugin/internal/handler"
	"github.com/usegiam/giam-traefik-plugin/internal/types"
	"github.com/usegiam/giam-traefik-plugin/pkg/grafana"
	"github.com/usegiam/giam-traefik-plugin/pkg/log"
)

const labelValuesEndpointPattern = `^/api/datasources/uid/([a-zA-Z0-9-]+)/resources/label/([\w]+)/values`

var labelValuesEndpointRegexExp = regexp.MustCompile(labelValuesEndpointPattern)

type LabelValuesHandler struct {
	service     loki.Service
	grafanaRepo grafana.Repo
	logger      *log.Logger
}

type LabelValuesHandlerDeps struct {
	Service     loki.Service
	GrafanaRepo grafana.Repo
	Logger      *log.Logger
}

func NewLabelValueHandler(deps *LabelValuesHandlerDeps) handler.Handler {
	return &LabelValuesHandler{service: deps.Service, grafanaRepo: deps.GrafanaRepo, logger: deps.Logger}
}

func (l *LabelValuesHandler) Match(req *http.Request) bool {
	pattern := regexp.MustCompile(labelValuesEndpointPattern)

	if !pattern.MatchString(req.RequestURI) {
		return false
	}

	datasourceType := req.Header.Get("X-Plugin-Id")

	return datasourceType == string(datasource.Loki)
}

func (l *LabelValuesHandler) Handle(rw http.ResponseWriter, req *http.Request, next http.Handler) {
	l.logger.Debug("instantiated a loki label values filter")

	matches := labelValuesEndpointRegexExp.FindStringSubmatch(req.RequestURI)

	w := &types.ResponseWriter{
		ResponseWriter: rw,
		Body:           &bytes.Buffer{},
		Status:         http.StatusOK,
	}

	grafanaSession, err := req.Cookie("grafana_session")
	if err != nil {
		http.Error(rw, "Forbidden", http.StatusForbidden)

		return
	}

	next.ServeHTTP(w, req)

	l.logger.Debugf("grafana response %s", string(w.Body.Bytes()))

	var grafanaResp grafana.LabelValuesReq

	err = json.Unmarshal(w.Body.Bytes(), &grafanaResp)
	if err != nil {
		http.Error(rw, "Internal server error", http.StatusPreconditionFailed)

		return
	}

	user, err := l.grafanaRepo.GetUser(grafanaSession.Value)
	if err != nil {
		l.logger.Debugf("user doesn't exists, err: %v", err)

		http.Error(rw, "User doesn't exits", http.StatusBadRequest)

		return
	}

	teams, err := l.grafanaRepo.GetUserTeams(grafanaSession.Value, user.ID)
	if err != nil {
		l.logger.Debugf("user doesn't have any team, err: %v", err)

		http.Error(rw, "User not assigned to any team", http.StatusBadRequest)

		return
	}

	resp, err := l.service.FilterLabelValues(&loki.FilterLabelValuesReq{
		User:  user,
		Teams: teams,
		Label: &loki.Label{
			Name:   matches[2], // Label name exists in second index of the regx
			Values: grafanaResp.Data,
		},
		Datasource: grafana.Datasource{UID: matches[1]}, // Uid exists in first index of the regx
	})
	if err != nil {
		l.logger.Debugf("unable to send loki filter label request to Giam, err: %v", err)

		http.Error(rw, "Unable to communicate with Giam service", http.StatusPreconditionFailed)

		return
	}

	grafanaResp.Data = resp.Data

	responseBody, err := json.Marshal(grafanaResp)
	if err != nil {
		http.Error(rw, "Error marshaling response", http.StatusInternalServerError)

		return
	}

	rw.Header().Set("Content-Length", strconv.Itoa(len(responseBody)))
	rw.WriteHeader(w.Status)
	rw.Write(responseBody)
}
