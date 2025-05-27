package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/usegiam/giam-traefik-plugin/internal/datasource/prometheus"
	"github.com/usegiam/giam-traefik-plugin/internal/datasource/prometheus/service"
	"github.com/usegiam/giam-traefik-plugin/pkg/grafana"
	"github.com/usegiam/giam-traefik-plugin/pkg/hash"
	"github.com/usegiam/giam-traefik-plugin/pkg/log"
	"github.com/usegiam/giam-traefik-plugin/pkg/testutil/assert"
	"github.com/usegiam/giam-traefik-plugin/pkg/testutil/mocks"
	"github.com/usegiam/giam-traefik-plugin/pkg/testutil/require"
)

func TestQueryHandler_Handle(t *testing.T) {
	tests := []struct {
		name               string
		payload            *grafana.QueryReq
		expectedBody       interface{}
		grafanaRepo        grafana.Repo
		hashSvc            hash.Service
		prometheusSvc      prometheus.Service
		expectedStatusCode int
	}{
		{
			name: "Test with multiple queries",
			payload: &grafana.QueryReq{
				Queries: []interface{}{
					map[string]interface{}{
						"expr": `http_status{cluster="customer1"}`,
						"datasource": map[string]interface{}{
							"uid": "dummyUID1",
						},
					},
					map[string]interface{}{
						"expr": `http_status{cluster="customer2", team="payment"}`,
						"datasource": map[string]interface{}{
							"uid": "dummyUID2",
						},
					},
				},
			},
			grafanaRepo: &grafana.MockRepo{
				User:  &grafana.User{ID: 1, Name: "user1"},
				Teams: []*grafana.Team{{ID: 1, Name: "team1"}},
				Err:   nil,
			},
			hashSvc: &hash.MockService{
				Hash: "dummyHash",
				Err:  nil,
			},
			prometheusSvc: &service.Mock{
				AuthorizedQueryResp: &prometheus.AuthorizedQueryResp{
					Queries: []interface{}{
						map[string]interface{}{
							"expr": `http_status{cluster="customer1", team=~"menu"}`,
							"datasource": map[string]interface{}{
								"uid": "dummyUID1",
							},
						},
						map[string]interface{}{
							"expr": `http_status{cluster="customer2", team="payment"}`,
							"datasource": map[string]interface{}{
								"uid": "dummyUID2",
							},
						},
					},
					Message:    "random message",
					StatusCode: http.StatusOK,
				},
				Error: nil,
			},
			expectedBody: &grafana.QueryReq{
				Queries: []interface{}{
					map[string]interface{}{
						"expr": `http_status{cluster="customer1", team=~"menu"}`,
						"datasource": map[string]interface{}{
							"uid": "dummyUID1",
						},
					},
					map[string]interface{}{
						"expr": `http_status{cluster="customer2", team="payment"}`,
						"datasource": map[string]interface{}{
							"uid": "dummyUID2",
						},
					},
				},
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "Test with single query",
			payload: &grafana.QueryReq{
				Queries: []interface{}{
					map[string]interface{}{
						"expr": `http_status{cluster="customer3"}`,
						"datasource": map[string]interface{}{
							"uid": "dummyUID3",
						},
					},
				},
			},
			grafanaRepo: &grafana.MockRepo{
				User:  &grafana.User{ID: 1, Name: "user1"},
				Teams: []*grafana.Team{{ID: 1, Name: "team1"}},
				Err:   nil,
			},
			hashSvc: &hash.MockService{
				Hash: "dummyHash",
				Err:  nil,
			},
			prometheusSvc: &service.Mock{
				AuthorizedQueryResp: &prometheus.AuthorizedQueryResp{
					Queries: []interface{}{
						map[string]interface{}{
							"expr": `http_status{cluster="customer3"}`,
							"datasource": map[string]interface{}{
								"uid": "dummyUID3",
							},
						},
					},
					Message:    "random message",
					StatusCode: http.StatusOK,
				},
				Error: nil,
			},
			expectedBody: &grafana.QueryReq{
				Queries: []interface{}{
					map[string]interface{}{
						"expr": `http_status{cluster="customer3"}`,
						"datasource": map[string]interface{}{
							"uid": "dummyUID3",
						},
					},
				},
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "Test with single query when there is no team assigned for the user",
			payload: &grafana.QueryReq{
				Queries: []interface{}{
					map[string]interface{}{
						"expr": `http_status{cluster="customer3"}`,
						"datasource": map[string]interface{}{
							"uid": "dummyUID3",
						},
					},
				},
			},
			grafanaRepo: &grafana.MockRepo{
				User:  &grafana.User{ID: 1, Name: "user1"},
				Teams: []*grafana.Team{},
				Err:   nil,
			},
			hashSvc: &hash.MockService{
				Hash: "dummyHash",
				Err:  nil,
			},
			prometheusSvc: &service.Mock{
				AuthorizedQueryResp: &prometheus.AuthorizedQueryResp{
					Message:    "No Team Assigned",
					StatusCode: http.StatusPreconditionFailed,
				},
				Error: nil,
			},
			expectedBody:       "No Team Assigned",
			expectedStatusCode: http.StatusPreconditionFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonPayload, err := json.Marshal(tt.payload)

			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/ds/query?ds_type=prometheus", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")

			cookie := &http.Cookie{
				Name:  "grafana_session",
				Value: "mocked_session_value",
			}
			req.AddCookie(cookie)

			rr := httptest.NewRecorder()
			handler := &QueryHandler{
				logger:        log.New("FATAL"),
				grafanaRepo:   tt.grafanaRepo,
				hashSvc:       tt.hashSvc,
				prometheusSvc: tt.prometheusSvc,
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

func TestQueryHandler_Match(t *testing.T) {
	type fields struct {
		svc prometheus.Service
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
			name:   "it should return true when the endpoint is for query",
			fields: fields{svc: &service.Mock{}},
			args: args{
				req: httptest.NewRequest(http.MethodPost, "/api/ds/query?ds_type=prometheus", nil),
			},
			want: true,
		},
		{
			name:   "it should return true when the endpoint is for query",
			fields: fields{svc: &service.Mock{}},
			args: args{
				req: httptest.NewRequest(http.MethodPost, "/api/ds/query?ds_type=loki", nil),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &QueryHandler{
				prometheusSvc: tt.fields.svc,
			}
			if got := l.Match(tt.args.req); got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}
