package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"regexp"

	"github.com/usegiam/giam-traefik-plugin/internal/authorization"
	"github.com/usegiam/giam-traefik-plugin/internal/handler"
	"github.com/usegiam/giam-traefik-plugin/pkg/grafana"
	"github.com/usegiam/giam-traefik-plugin/pkg/log"
)

// datasourceQueryEndpointPattern this endpoint is for grafana querying. We don't include the base url in the pattern because
// in proxy it will be without the base url.
const datasourceQueryEndpointPattern = "^/api/ds/query"

var datasourceQueryEndpointRegexExp = regexp.MustCompile(datasourceQueryEndpointPattern)

type DatasourceHandler struct {
	logger      *log.Logger
	grafanaRepo grafana.Repo
	service     authorization.Service
}

type DatasourceHandlerDeps struct {
	Logger      *log.Logger
	GrafanaRepo grafana.Repo
	Service     authorization.Service
}

func NewDatasourceHandler(deps *DatasourceHandlerDeps) handler.Handler {
	return &DatasourceHandler{
		logger:      deps.Logger,
		service:     deps.Service,
		grafanaRepo: deps.GrafanaRepo,
	}
}

func (d *DatasourceHandler) Match(req *http.Request) bool {
	return datasourceQueryEndpointRegexExp.MatchString(req.RequestURI)
}

func (d *DatasourceHandler) Handle(rw http.ResponseWriter, req *http.Request, next http.Handler) {
	d.logger.Debug("instantiated a datasource authorize")

	rawBody, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(rw, "Unable to read request body", http.StatusBadRequest)

		return
	}

	defer req.Body.Close()

	var queryReq grafana.QueryReq
	if err := json.Unmarshal(rawBody, &queryReq); err != nil {
		http.Error(rw, "Invalid JSON", http.StatusBadRequest)

		return
	}

	grafanaSession, err := req.Cookie("grafana_session")
	if err != nil {
		http.Error(rw, "Forbidden", http.StatusForbidden)

		return
	}

	user, err := d.grafanaRepo.GetUser(grafanaSession.Value)
	if err != nil {
		d.logger.Debugf("user doesn't exists, err: %v", err)

		http.Error(rw, "User doesn't exits", http.StatusBadRequest)

		return
	}

	teams, err := d.grafanaRepo.GetUserTeams(grafanaSession.Value, user.ID)
	if err != nil {
		d.logger.Debugf("user doesn't have any team, err: %v", err)

		http.Error(rw, "User not assigned to any team", http.StatusBadRequest)

		return
	}

	resp, err := d.service.AuthorizeQuery(&authorization.AuthorizeDatasourceReq{
		User:    user,
		Teams:   teams,
		Queries: queryReq.Queries,
	})
	if err != nil {
		d.logger.Debugf("unable to send authorize request to Giam, err: %v", err)

		http.Error(rw, "Unable to communicate with Giam service", http.StatusPreconditionFailed)

		return
	}

	if resp.StatusCode != http.StatusOK {
		d.logger.Debugf("user is not authorized to given datasource")

		http.Error(rw, resp.Message, resp.StatusCode)

		return
	}

	req.Body = io.NopCloser(bytes.NewBuffer(rawBody))
	req.ContentLength = int64(len(rawBody))

	next.ServeHTTP(rw, req)
}
