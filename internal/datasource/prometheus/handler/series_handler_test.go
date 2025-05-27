package handler

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/usegiam/giam-traefik-plugin/internal/datasource/prometheus"
	"github.com/usegiam/giam-traefik-plugin/internal/datasource/prometheus/service"
	"github.com/usegiam/giam-traefik-plugin/pkg/grafana"
	"github.com/usegiam/giam-traefik-plugin/pkg/log"
	"github.com/usegiam/giam-traefik-plugin/pkg/testutil/assert"
	"github.com/usegiam/giam-traefik-plugin/pkg/testutil/mocks"
	"github.com/usegiam/giam-traefik-plugin/pkg/testutil/require"
)

func TestSeriesHandler_Handle(t *testing.T) {
	tests := []struct {
		name               string
		payload            *grafana.SeriesReq
		mockedResponse     string
		expectedBody       grafana.SeriesReq
		service            prometheus.Service
		grafanaRepo        grafana.Repo
		headers            http.Header
		expectedStatusCode int
		compressedResponse bool
	}{
		{
			name: "It should return the filter label values",
			payload: &grafana.SeriesReq{
				Series: []map[string]string{
					{
						"__name__":  "prometheus_operator_build_info",
						"branch":    "refs/tags/v0.66.0",
						"customer":  "customer1-staging",
						"container": "kube-prometheus-stack",
					},
					{
						"__name__":  "prometheus_operator_build_info",
						"branch":    "refs/tags/v0.66.0",
						"customer":  "customer1-production",
						"container": "kube-prometheus-stack",
					},
					{
						"__name__":  "prometheus_operator_build_info",
						"branch":    "refs/tags/v0.66.0",
						"customer":  "customer2-staging",
						"container": "kube-prometheus-stack",
					},
					{
						"__name__":  "prometheus_operator_build_info",
						"branch":    "refs/tags/v0.66.0",
						"customer":  "customer2-production",
						"container": "kube-prometheus-stack",
					},
				},
				Status: "success",
			},
			mockedResponse: `{
				"data": [
					{
						"__name__": "prometheus_operator_build_info",
						"branch": "refs/tags/v0.66.0",
						"customer": "customer1-staging",
						"container": "kube-prometheus-stack"
					},
					{
						"__name__": "prometheus_operator_build_info",
						"branch": "refs/tags/v0.66.0",
						"customer": "customer1-production",
						"container": "kube-prometheus-stack"
					},
					{
						"__name__": "prometheus_operator_build_info",
						"branch": "refs/tags/v0.66.0",
						"customer": "customer2-staging",
						"container": "kube-prometheus-stack"
					},
					{
						"__name__": "prometheus_operator_build_info",
						"branch": "refs/tags/v0.66.0",
						"customer": "customer2-production",
						"container": "kube-prometheus-stack"
					}
				]}`,
			expectedBody: grafana.SeriesReq{
				Series: []map[string]string{
					{
						"__name__":  "prometheus_operator_build_info",
						"branch":    "refs/tags/v0.66.0",
						"customer":  "customer1-staging",
						"container": "kube-prometheus-stack",
					},
					{
						"__name__":  "prometheus_operator_build_info",
						"branch":    "refs/tags/v0.66.0",
						"customer":  "customer1-production",
						"container": "kube-prometheus-stack",
					},
				},
				Status: "success",
			},
			service: &service.Mock{
				FilterSeriesResp: &prometheus.FilterSeriesResp{
					Data: []map[string]string{
						{
							"__name__":  "prometheus_operator_build_info",
							"branch":    "refs/tags/v0.66.0",
							"customer":  "customer1-staging",
							"container": "kube-prometheus-stack",
						},
						{
							"__name__":  "prometheus_operator_build_info",
							"branch":    "refs/tags/v0.66.0",
							"customer":  "customer1-production",
							"container": "kube-prometheus-stack",
						},
					},
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
			name: "It should return an error for invalid JSON response",
			payload: &grafana.SeriesReq{
				Series: []map[string]string{},
				Status: "success",
			},
			mockedResponse:     `{ "invalid JSON"`,
			expectedBody:       grafana.SeriesReq{},
			service:            &service.Mock{},
			expectedStatusCode: http.StatusPreconditionFailed,
		},
		{
			name: "It should return an empty filtered series for empty upstream data",
			payload: &grafana.SeriesReq{
				Series: []map[string]string{},
				Status: "success",
			},
			mockedResponse: `{
				"data": [],
				"status": "success"
			}`,
			expectedBody: grafana.SeriesReq{
				Series: []map[string]string{},
				Status: "success",
			},
			service: &service.Mock{
				FilterSeriesResp: &prometheus.FilterSeriesResp{
					Data: []map[string]string{},
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
			name: "It should handle gzip-compressed response",
			payload: &grafana.SeriesReq{
				Series: []map[string]string{
					{
						"__name__":  "prometheus_operator_build_info",
						"branch":    "refs/tags/v0.66.0",
						"customer":  "customer1-staging",
						"container": "kube-prometheus-stack",
					},
					{
						"__name__":  "prometheus_operator_build_info",
						"branch":    "refs/tags/v0.66.0",
						"customer":  "customer1-production",
						"container": "kube-prometheus-stack",
					},
					{
						"__name__":  "prometheus_operator_build_info",
						"branch":    "refs/tags/v0.66.0",
						"customer":  "customer2-staging",
						"container": "kube-prometheus-stack",
					},
					{
						"__name__":  "prometheus_operator_build_info",
						"branch":    "refs/tags/v0.66.0",
						"customer":  "customer2-production",
						"container": "kube-prometheus-stack",
					},
				},
				Status: "success",
			},
			mockedResponse: `{
				"data": [
					{
						"__name__": "prometheus_operator_build_info",
						"branch": "refs/tags/v0.66.0",
						"customer": "customer1-staging",
						"container": "kube-prometheus-stack"
					},
					{
						"__name__": "prometheus_operator_build_info",
						"branch": "refs/tags/v0.66.0",
						"customer": "customer1-production",
						"container": "kube-prometheus-stack"
					},
					{
						"__name__": "prometheus_operator_build_info",
						"branch": "refs/tags/v0.66.0",
						"customer": "customer2-staging",
						"container": "kube-prometheus-stack"
					},
					{
						"__name__": "prometheus_operator_build_info",
						"branch": "refs/tags/v0.66.0",
						"customer": "customer2-production",
						"container": "kube-prometheus-stack"
					}
				]}`,
			expectedBody: grafana.SeriesReq{
				Series: []map[string]string{
					{
						"__name__":  "prometheus_operator_build_info",
						"branch":    "refs/tags/v0.66.0",
						"customer":  "customer1-staging",
						"container": "kube-prometheus-stack",
					},
					{
						"__name__":  "prometheus_operator_build_info",
						"branch":    "refs/tags/v0.66.0",
						"customer":  "customer1-production",
						"container": "kube-prometheus-stack",
					},
				},
				Status: "success",
			},
			service: &service.Mock{
				FilterSeriesResp: &prometheus.FilterSeriesResp{
					Data: []map[string]string{
						{
							"__name__":  "prometheus_operator_build_info",
							"branch":    "refs/tags/v0.66.0",
							"customer":  "customer1-staging",
							"container": "kube-prometheus-stack",
						},
						{
							"__name__":  "prometheus_operator_build_info",
							"branch":    "refs/tags/v0.66.0",
							"customer":  "customer1-production",
							"container": "kube-prometheus-stack",
						},
					},
				},
			},
			grafanaRepo: &grafana.MockRepo{
				User:  &grafana.User{ID: 1, Name: "user1"},
				Teams: []*grafana.Team{{ID: 1, Name: "team1"}},
				Err:   nil,
			},
			headers:            http.Header{"Content-Encoding": []string{"gzip"}},
			expectedStatusCode: http.StatusOK,
			compressedResponse: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonPayload, err := json.Marshal(tt.payload)

			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/datasources/uid/P0dfd3df3dfd/resources/api/v1/series", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")

			cookie := &http.Cookie{
				Name:  "grafana_session",
				Value: "mocked_session_value",
			}
			req.AddCookie(cookie)

			rr := httptest.NewRecorder()

			mockResponseBody := []byte(tt.mockedResponse)

			if tt.compressedResponse {
				var buf bytes.Buffer
				gz := gzip.NewWriter(&buf)
				_, err = gz.Write(mockResponseBody)
				require.NoError(t, err)
				require.NoError(t, gz.Close())
				mockResponseBody = buf.Bytes()
			}

			handler := &SeriesHandler{
				logger:      log.New("FATAL"),
				service:     tt.service,
				grafanaRepo: tt.grafanaRepo,
			}

			handler.Handle(rr, req, &mocks.NextHandler{
				RespBody:  mockResponseBody,
				HeaderMap: tt.headers,
			})

			assert.Equal(t, tt.expectedStatusCode, rr.Code)

			if tt.expectedStatusCode == http.StatusOK {
				var actualResponse grafana.SeriesReq

				if rr.Header().Get("Content-Encoding") == "gzip" {
					gzReader, err := gzip.NewReader(rr.Body)

					require.NoError(t, err)
					defer gzReader.Close()

					err = json.NewDecoder(gzReader).Decode(&actualResponse)

					require.NoError(t, err)
				} else {
					err = json.NewDecoder(rr.Body).Decode(&actualResponse)

					require.NoError(t, err)
				}

				assert.CompareJson(t, tt.expectedBody.Series, actualResponse.Series)
			} else {
				wantedBody := strings.TrimSpace(rr.Body.String())
				expectedBody := "Internal server error"

				assert.Equal(t, expectedBody, wantedBody)
			}
		})
	}
}

func TestSeriesHandler_Match(t *testing.T) {
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
			name:   "it should return true when the endpoint is for series",
			fields: fields{svc: &service.Mock{}},
			args: args{
				req: httptest.NewRequest(http.MethodPost, "/api/datasources/uid/d4005dd5-6a69-4d37-aaca-7a5c7975bd98/resources/api/v1/series?start=1729843380&end=1729847040", nil),
			},
			want: true,
		},
		{
			name:   "it should return false when the endpoint is not for series",
			fields: fields{svc: &service.Mock{}},
			args: args{
				req: httptest.NewRequest(http.MethodPost, "/api/ds/query?ds_type=prometheus", nil),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &SeriesHandler{
				service: tt.fields.svc,
			}
			if got := l.Match(tt.args.req); got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}
