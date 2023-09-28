package counter

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vorotislav/alert-service/internal/http/middlewares"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

type mockStorage struct {
}

func (m *mockStorage) UpdateCounter(_ string, _ int64) error {
	return nil
}

func (m *mockStorage) GetCounterValue(name string) (int64, error) {
	if name == "PollCount" {
		return 15, nil
	}

	return 0, fmt.Errorf("some error")
}

func (m *mockStorage) AllCounterMetrics() ([]byte, error) {
	return nil, nil
}

func TestHandler_Value(t *testing.T) {
	log, err := zap.NewDevelopment()
	require.NoError(t, err)

	tests := []struct {
		name           string
		giveMethod     string
		givePath       string
		wantStatusCode int
		wantValue      int64
	}{
		{
			name:           "success",
			giveMethod:     http.MethodGet,
			givePath:       "/value/counter/PollCount",
			wantStatusCode: http.StatusOK,
			wantValue:      15,
		},
		{
			name:           "not allowed",
			giveMethod:     http.MethodPost,
			givePath:       "/value/counter/PollCount",
			wantStatusCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := chi.NewRouter()
			h := &Handler{
				log: log,
				s:   &mockStorage{},
			}
			r.Use(middlewares.New(log))

			r.Route("/value", func(r chi.Router) {
				r.Route("/counter", func(r chi.Router) {
					r.Route("/{metricName}", func(r chi.Router) {
						r.Get("/", h.Value)
					})
				})

				r.Route("/{metricType}", func(r chi.Router) {
					r.Get("/{metricName}", func(writer http.ResponseWriter, request *http.Request) {
						http.Error(writer, "", http.StatusBadRequest)
					})
				})

			})

			server := httptest.NewServer(r)
			defer server.Close()

			request, err := http.NewRequest(tc.giveMethod, server.URL+tc.givePath, nil)
			require.NoError(t, err)

			res, err := server.Client().Do(request)
			require.NoError(t, err)
			defer res.Body.Close()
			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			if len(body) > 0 {
				value, err := strconv.Atoi(string(body))
				require.NoError(t, err)
				assert.Equal(t, tc.wantValue, int64(value))
			}

			assert.Equal(t, tc.wantStatusCode, res.StatusCode)
		})
	}
}

func TestHandler_Update(t *testing.T) {
	log, err := zap.NewDevelopment()
	require.NoError(t, err)

	tests := []struct {
		name           string
		giveMethod     string
		givePath       string
		wantStatusCode int
	}{
		{
			name:           "success",
			givePath:       "/update/counter/someMetric/1",
			giveMethod:     http.MethodPost,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "method not allowed",
			giveMethod:     http.MethodGet,
			givePath:       "/update/counter/someMetric/1",
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:           "method not found",
			giveMethod:     http.MethodPost,
			givePath:       "/update/counter/1",
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "bad request",
			giveMethod:     http.MethodPost,
			givePath:       "/update/counter/someMetric/metric",
			wantStatusCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			h := &Handler{
				log: log,
				s:   &mockStorage{},
			}

			r.Use(middlewares.New(log))

			r.Route("/update", func(r chi.Router) {
				r.Route("/counter", func(r chi.Router) {
					r.Route("/{metricName}", func(r chi.Router) {
						r.Post("/{metricValue}", h.Update)
					})
				})

				r.Route("/{metricType}", func(r chi.Router) {
					r.Route("/{metricName}", func(r chi.Router) {
						r.Post("/{metricValue}", func(writer http.ResponseWriter, request *http.Request) {
							http.Error(writer, "", http.StatusBadRequest)
						})
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
