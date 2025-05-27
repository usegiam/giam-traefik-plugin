package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"regexp"

	"github.com/usegiam/giam-traefik-plugin/internal/datasource"
	"github.com/usegiam/giam-traefik-plugin/internal/datasource/loki"
	"github.com/usegiam/giam-traefik-plugin/internal/handler"
	"github.com/usegiam/giam-traefik-plugin/pkg/grafana"
	"github.com/usegiam/giam-traefik-plugin/pkg/log"
)

// queryEndpointPattern this endpoint is for grafana querying. We don't include the base url in the pattern because
// in proxy it will be without the base url.
const queryEndpointPattern = "^/api/ds/query"

var queryEndpointRegexExp = regexp.MustCompile(queryEndpointPattern)

type QueryHandler struct {
	service     loki.Service
	grafanaRepo grafana.Repo
	logger      *log.Logger
}

type QueryHandlerDeps struct {
	Service     loki.Service
	GrafanaRepo grafana.Repo
	Logger      *log.Logger
}

func NewQueryHandler(deps *QueryHandlerDeps) handler.Handler {
	return &QueryHandler{service: deps.Service, grafanaRepo: deps.GrafanaRepo, logger: deps.Logger}
}

func (l *QueryHandler) Match(req *http.Request) bool {
	if !queryEndpointRegexExp.MatchString(req.RequestURI) {
		return false
	}

	datasourceType := req.URL.Query().Get("ds_type")

	return datasourceType == string(datasource.Loki)
}

func (l *QueryHandler) Handle(rw http.ResponseWriter, req *http.Request, next http.Handler) {
	l.logger.Debug("instantiated a loki query authorize")

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(rw, "Unable to read request body", http.StatusBadRequest)

		return
	}

	defer req.Body.Close()

	var queryReq grafana.QueryReq
	if err := json.Unmarshal(body, &queryReq); err != nil {
		http.Error(rw, "Invalid JSON", http.StatusBadRequest)

		return
	}

	grafanaSession, err := req.Cookie("grafana_session")
	if err != nil {
		http.Error(rw, "Forbidden", http.StatusForbidden)

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

	resp, err := l.service.AuthorizeQuery(&loki.AuthorizeQueryReq{
		User:    user,
		Teams:   teams,
		Queries: queryReq.Queries,
	})
	if err != nil {
		l.logger.Debugf("unable to send loki authorize query request to Giam, err: %v", err)

		http.Error(rw, "Unable to communicate with Giam service", http.StatusPreconditionFailed)

		return
	}

	if resp.StatusCode != http.StatusOK {
		http.Error(rw, resp.Message, resp.StatusCode)

		return
	}

	l.logger.Debugf("original queries: %v", queryReq.Queries)

	queryReq.Queries = resp.Queries

	l.logger.Debugf("replaced queries: %v", resp.Queries)

	updatedBody, err := json.Marshal(queryReq)
	if err != nil {
		http.Error(rw, "Error marshaling JSON", http.StatusOK)

		return
	}

	l.logger.Debugf("new request body: %s", string(updatedBody))

	req.Body = io.NopCloser(bytes.NewBuffer(updatedBody))
	req.ContentLength = int64(len(updatedBody))

	next.ServeHTTP(rw, req)
}
