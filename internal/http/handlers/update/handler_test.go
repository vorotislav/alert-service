package update

import (
	"bytes"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vorotislav/alert-service/internal/http/middlewares"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubStorage struct {
}

func (m *stubStorage) UpdateCounter(_ string, _ int64) (int64, error) {
	return 0, nil
}

func (m *stubStorage) UpdateGauge(_ string, _ float64) (float64, error) {
	return 0, nil
}

func TestHandler_Update(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)

	tests := []struct {
		name           string
		giveMethod     string
		givePath       string
		wantStatusCode int
	}{
		{
			name:           "success counter",
			givePath:       "/update/counter/someMetric/1",
			giveMethod:     http.MethodPost,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "method not allowed counter",
			giveMethod:     http.MethodGet,
			givePath:       "/update/counter/someMetric/1",
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:           "method not found counter",
			giveMethod:     http.MethodPost,
			givePath:       "/update/counter/1",
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "bad request counter",
			giveMethod:     http.MethodPost,
			givePath:       "/update/counter/someMetric/metric",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "success gauge",
			giveMethod:     http.MethodPost,
			givePath:       "/update/gauge/someMetric/1",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "method not allowed gauge",
			giveMethod:     http.MethodGet,
			givePath:       "/update/gauge/someMetric/1",
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:           "method not found gauge",
			giveMethod:     http.MethodPost,
			givePath:       "/update/gauge/1",
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "bad request gauge",
			giveMethod:     http.MethodPost,
			givePath:       "/update/gauge/someMetric/metric",
			wantStatusCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := chi.NewRouter()
			h := &Handler{
				log:     log,
				storage: &stubStorage{},
			}

			r.Use(middlewares.New(log))

			r.Route("/update", func(r chi.Router) {
				r.Route("/{metricType}", func(r chi.Router) {
					r.Route("/{metricName}", func(r chi.Router) {
						r.Post("/{metricValue}", h.Update)
					})
				})
			})
			server := httptest.NewServer(r)
			defer server.Close()

			request, err := http.NewRequest(tt.giveMethod, server.URL+tt.givePath, nil)
			require.NoError(t, err)

			res, err := server.Client().Do(request)
			require.NoError(t, err)
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatusCode, res.StatusCode)
		})
	}
}

func TestHandler_UpdateJSON(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)

	tests := []struct {
		name           string
		giveMethod     string
		giveBody       []byte
		wantStatusCode int
	}{
		{
			name:           "success counter",
			giveBody:       []byte(`{"id":"some counter", "type":"counter", "delta":1}`),
			giveMethod:     http.MethodPost,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "method not allowed counter",
			giveMethod:     http.MethodGet,
			giveBody:       []byte(`{"id":"some counter", "type":"counter", "delta":1}`),
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:           "bad request counter",
			giveMethod:     http.MethodPost,
			giveBody:       []byte(`{"name":"some counter", "type":"azaza", "delta":1}`),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "success gauge",
			giveMethod:     http.MethodPost,
			giveBody:       []byte(`{"id":"some metric", "type":"gauge", "value":1}`),
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "method not allowed gauge",
			giveMethod:     http.MethodGet,
			giveBody:       []byte(`{"id":"some metric", "type":"gauge", "value":1}`),
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:           "bad request gauge",
			giveMethod:     http.MethodPost,
			giveBody:       []byte(`{"id":"some metric", "type":"gauge"}`),
			wantStatusCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			h := &Handler{
				log:     log,
				storage: &stubStorage{},
			}

			r.Use(middlewares.New(log))

			r.Route("/update", func(r chi.Router) {
				r.Post("/", h.UpdateJSON)
			})

			server := httptest.NewServer(r)
			defer server.Close()

			request, err := http.NewRequest(tt.giveMethod, server.URL+"/update", bytes.NewBuffer(tt.giveBody))
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")

			res, err := server.Client().Do(request)
			require.NoError(t, err)
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatusCode, res.StatusCode)
		})
	}
}
