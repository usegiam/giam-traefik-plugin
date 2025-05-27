package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/usegiam/giam-traefik-plugin/internal/datasource/loki"
	"github.com/usegiam/giam-traefik-plugin/internal/datasource/loki/service"
	"github.com/usegiam/giam-traefik-plugin/pkg/grafana"
	"github.com/usegiam/giam-traefik-plugin/pkg/log"
	"github.com/usegiam/giam-traefik-plugin/pkg/testutil/assert"
	"github.com/usegiam/giam-traefik-plugin/pkg/testutil/mocks"
	"github.com/usegiam/giam-traefik-plugin/pkg/testutil/require"
)

func TestLabelValuesHandler_Handle(t *testing.T) {
	tests := []struct {
		name               string
		payload            *grafana.LabelValuesReq
		expectedBody       interface{}
		service            loki.Service
		grafanaRepo        grafana.Repo
		expectedStatusCode int
	}{
		{
			name: "It should return the filtered label values",
			payload: &grafana.LabelValuesReq{
				Data:   []string{"customer1", "customer2", "customer3"},
				Status: "success",
			},
			expectedBody: grafana.LabelValuesReq{
				Data:   []string{"customer1", "customer3"},
				Status: "success",
			},
			service: &service.Mock{
				FilterLabelValuesResp: &loki.FilterLabelValuesResp{
					Data: []string{"customer1", "customer3"},
				},
			},
			grafanaRepo: &grafana.MockRepo{
				User:  &grafana.User{ID: 1, Name: "user1"},
				Teams: []*grafana.Team{{ID: 1, Name: "team1"}},
				Err:   nil,
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonPayload, err := json.Marshal(tt.payload)

			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/datasources/uid/P0dfd3df3dfd/resources/label/cluster/values", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")

			cookie := &http.Cookie{
				Name:  "grafana_session",
				Value: "mocked_session_value",
			}
			req.AddCookie(cookie)

			rr := httptest.NewRecorder()
			handler := &LabelValuesHandler{
				logger:      log.New("FATAL"),
				service:     tt.service,
				grafanaRepo: tt.grafanaRepo,
			}

			handler.Handle(rr, req, &mocks.NextHandler{
				RespBody: []byte(`{"data": ["customer1", "customer2"], "status": "success"}`),
			})

			require.NoError(t, err)
			assert.Equal(t, rr.Code, tt.expectedStatusCode)

			if expectedCastedBody, matched := tt.expectedBody.(grafana.LabelValuesReq); matched {
				var actualRequestBody grafana.LabelValuesReq

				modifiedBody, err := io.ReadAll(rr.Body)
				err = json.Unmarshal(modifiedBody, &actualRequestBody)

				require.NoError(t, err)

				contentLength, err := strconv.ParseInt(rr.Header().Get("Content-Length"), 10, 64)

				require.NoError(t, err)

				assert.Equal(t, contentLength, int64(len(modifiedBody)))
				assert.CompareJson(t, expectedCastedBody.Data, actualRequestBody.Data)
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

func TestLabelValuesHandler_Match(t *testing.T) {
	type fields struct {
		svc loki.Service
	}

	type args struct {
		req func() *http.Request
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "it should return true when the endpoint is for label values",
			fields: fields{svc: &service.Mock{}},
			args: args{
				req: func() *http.Request {
					req := httptest.NewRequest(http.MethodPost, "/api/datasources/uid/d4005dd5-6a69-4d37-aaca-7a5c7975bd98/resources/label/cluster/values", nil)

					req.Header.Set("X-Plugin-Id", "loki")

					return req
				},
			},
			want: true,
		},
		{
			name:   "it should return false when the endpoint is not for label values",
			fields: fields{svc: &service.Mock{}},
			args: args{
				req: func() *http.Request {
					return httptest.NewRequest(http.MethodPost, "/api/ds/query?ds_type=prometheus", nil)
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &LabelValuesHandler{
				service: tt.fields.svc,
			}
			if got := l.Match(tt.args.req()); got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}
