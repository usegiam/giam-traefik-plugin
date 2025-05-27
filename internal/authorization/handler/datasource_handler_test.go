package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/usegiam/giam-traefik-plugin/internal/authorization"
	"github.com/usegiam/giam-traefik-plugin/internal/authorization/service"
	"github.com/usegiam/giam-traefik-plugin/pkg/grafana"
	"github.com/usegiam/giam-traefik-plugin/pkg/log"
	"github.com/usegiam/giam-traefik-plugin/pkg/testutil/assert"
	"github.com/usegiam/giam-traefik-plugin/pkg/testutil/mocks"
	"github.com/usegiam/giam-traefik-plugin/pkg/testutil/require"
)

func TestDatasourceHandler_Handle(t *testing.T) {
	tests := []struct {
		name               string
		payload            *grafana.QueryReq
		expectedBody       interface{}
		service            authorization.Service
		grafanaRepo        grafana.Repo
		expectedStatusCode int
	}{
		{
			name: "Test with multiple queries",
			payload: &grafana.QueryReq{
				Queries: []interface{}{
					map[string]interface{}{
						"expr": `{cluster="customer1"}`,
					},
					map[string]interface{}{
						"expr": `{cluster="customer2", team="payment"}`,
					},
				},
			},
			expectedBody: &grafana.QueryReq{
				Queries: []interface{}{
					map[string]interface{}{
						"expr": `{cluster="customer1"}`,
					},
					map[string]interface{}{
						"expr": `{cluster="customer2", team="payment"}`,
					},
				},
			},
			service: &service.Mock{
				AuthorizeDatasourceResp: &authorization.AuthorizeDatasourceResp{
					Message:    "random message",
					StatusCode: http.StatusOK,
				},
			},
			grafanaRepo: &grafana.MockRepo{
				User:  &grafana.User{ID: 1, Name: "user1"},
				Teams: []*grafana.Team{{ID: 1, Name: "team1"}},
				Err:   nil,
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "Test with single query",
			payload: &grafana.QueryReq{
				Queries: []interface{}{
					map[string]interface{}{
						"expr": `{cluster="customer3"}`,
					},
				},
			},
			expectedBody: &grafana.QueryReq{
				Queries: []interface{}{
					map[string]interface{}{
						"expr": `{cluster="customer3"}`,
					},
				},
			},
			service: &service.Mock{
				AuthorizeDatasourceResp: &authorization.AuthorizeDatasourceResp{
					Message:    "random message",
					StatusCode: http.StatusOK,
				},
			},
			grafanaRepo: &grafana.MockRepo{
				User:  &grafana.User{ID: 1, Name: "user1"},
				Teams: []*grafana.Team{{ID: 1, Name: "team1"}},
				Err:   nil,
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "Test with single query when there is no team assigned for the user",
			payload: &grafana.QueryReq{
				Queries: []interface{}{
					map[string]interface{}{
						"expr": `{cluster="customer3"}`,
					},
				},
			},
			expectedBody: "No Team Assigned",
			service: &service.Mock{
				AuthorizeDatasourceResp: &authorization.AuthorizeDatasourceResp{
					Message:    "No Team Assigned",
					StatusCode: http.StatusPreconditionFailed,
				},
			},
			grafanaRepo: &grafana.MockRepo{
				User:  &grafana.User{ID: 1, Name: "user1"},
				Teams: []*grafana.Team{},
				Err:   nil,
			},
			expectedStatusCode: http.StatusPreconditionFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonPayload, err := json.Marshal(tt.payload)

			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/ds/query", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")

			cookie := &http.Cookie{
				Name:  "grafana_session",
				Value: "mocked_session_value",
			}
			req.AddCookie(cookie)

			rr := httptest.NewRecorder()
			handler := &DatasourceHandler{
				logger:      log.New("FATAL"),
				service:     tt.service,
				grafanaRepo: tt.grafanaRepo,
			}

			handler.Handle(rr, req, &mocks.NextHandler{})

			require.NoError(t, err)
			assert.Equal(t, rr.Code, tt.expectedStatusCode)

			if expectedCastedBody, matched := tt.expectedBody.(*grafana.QueryReq); matched {
				var actualRequestBody grafana.QueryReq

				modifiedBody, err := io.ReadAll(req.Body)
				err = json.Unmarshal(modifiedBody, &actualRequestBody)

				require.NoError(t, err)
				assert.Equal(t, req.ContentLength, int64(len(modifiedBody)))
				assert.CompareJson(t, expectedCastedBody.Queries, actualRequestBody.Queries)
			} else {
				wantedBody := strings.TrimSpace(rr.Body.String())
				expectedCastedBody := strings.TrimSpace(tt.expectedBody.(string))

				if expectedCastedBody != wantedBody {
					t.Errorf("Match() = %v, want = %v", expectedCastedBody, wantedBody)
				}
			}
		})
	}
}

func TestDatasourceHandler_Match(t *testing.T) {
	type fields struct {
		svc authorization.Service
	}

	type args struct {
		req *http.Request
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "it should return true when the endpoint is for datasource",
			fields: fields{svc: &service.Mock{}},
			args: args{
				req: httptest.NewRequest(http.MethodPost, "/api/ds/query", nil),
			},
			want: true,
		},
		{
			name:   "it should return false when the endpoint is not for datasource",
			fields: fields{svc: &service.Mock{}},
			args: args{
				req: httptest.NewRequest(http.MethodPost, "/api/ds/foo", nil),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &DatasourceHandler{
				service: tt.fields.svc,
			}
			if got := l.Match(tt.args.req); got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}
