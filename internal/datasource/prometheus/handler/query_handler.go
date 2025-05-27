package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"regexp"

	"github.com/usegiam/giam-traefik-plugin/internal/datasource"
	"github.com/usegiam/giam-traefik-plugin/internal/datasource/prometheus"
	"github.com/usegiam/giam-traefik-plugin/internal/handler"
	"github.com/usegiam/giam-traefik-plugin/pkg/grafana"
	"github.com/usegiam/giam-traefik-plugin/pkg/hash"
	"github.com/usegiam/giam-traefik-plugin/pkg/log"
)

// queryEndpointPattern this endpoint is for grafana querying. We don't include the base url in the pattern because
// in proxy it will be without the base url.
const queryEndpointPattern = "^/api/ds/query"

var queryEndpointRegexExp = regexp.MustCompile(queryEndpointPattern)

type QueryHandler struct {
	logger         *log.Logger
	grafanaRepo    grafana.Repo
	hashSvc        hash.Service
	prometheusRepo prometheus.Repo
	prometheusSvc  prometheus.Service
}

type QueryHandlerDeps struct {
	Logger         *log.Logger
	GrafanaRepo    grafana.Repo
	HashSvc        hash.Service
	PrometheusRepo prometheus.Repo
	PrometheusSvc  prometheus.Service
}

func NewQueryHandler(deps *QueryHandlerDeps) handler.Handler {
	return &QueryHandler{
		logger:         deps.Logger,
		hashSvc:        deps.HashSvc,
		grafanaRepo:    deps.GrafanaRepo,
		prometheusSvc:  deps.PrometheusSvc,
		prometheusRepo: deps.PrometheusRepo,
	}
}

func (l *QueryHandler) Match(req *http.Request) bool {
	if !queryEndpointRegexExp.MatchString(req.RequestURI) {
		return false
	}

	datasourceType := req.URL.Query().Get("ds_type")

	return datasourceType == string(datasource.Prometheus)
}

func (l *QueryHandler) Handle(rw http.ResponseWriter, req *http.Request, next http.Handler) {
	l.logger.Debug("instantiated a prometheus query authorize")

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
		http.Error(rw, "User doesn't exits", http.StatusBadRequest)

		return
	}

	teams, err := l.grafanaRepo.GetUserTeams(grafanaSession.Value, user.ID)
	if err != nil {
		l.logger.Debugf("user doesn't have any team, err: %v", err)

		http.Error(rw, "User not assigned to any team", http.StatusBadRequest)

		return
	}

	resp, err := l.prometheusSvc.AuthorizeQuery(&prometheus.AuthorizeQueryReq{
		User:    user,
		Teams:   teams,
		Queries: queryReq.Queries,
	})
	if err != nil {
		l.logger.Debugf("unable to send prometheus authorized query request to Giam, err: %v", err)

		http.Error(rw, "Unable to communicate with Giam service", http.StatusPreconditionFailed)

		return
	}

	if resp.StatusCode != http.StatusOK {
		http.Error(rw, resp.Message, resp.StatusCode)

		return
	}

	queryReq.Queries = resp.Queries

	updatedBody, err := json.Marshal(queryReq)
	if err != nil {
		http.Error(rw, "Error marshaling JSON", http.StatusOK)

		return
	}

	req.Body = io.NopCloser(bytes.NewBuffer(updatedBody))
	req.ContentLength = int64(len(updatedBody))

	next.ServeHTTP(rw, req)
}
