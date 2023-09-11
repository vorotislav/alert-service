package gauge

import (
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

type mockStorage struct {
}

func (m *mockStorage) UpdateGauge(name string, value float64) error {
	return nil
}

func (m *mockStorage) GetGaugeValue(name string) (float64, error) {
	if name == "mymetric" {
		return 11.1, nil
	}
	return 0, nil
}

func (m *mockStorage) AllGaugeMetrics() ([]byte, error) {
	return nil, nil
}

func TestHandler_Value(t *testing.T) {
	tests := []struct {
		name           string
		giveMethod     string
		givePath       string
		wantStatusCode int
		wantValue      float64
	}{
		{
			name:           "success",
			giveMethod:     http.MethodGet,
			givePath:       "/value/gauge/mymetric",
			wantStatusCode: http.StatusOK,
			wantValue:      11.1,
		},
		{
			name:           "not allowed",
			giveMethod:     http.MethodPost,
			givePath:       "/value/gauge/mymetric",
			wantStatusCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := chi.NewRouter()
			h := &Handler{
				Storage: &mockStorage{},
			}

			r.Route("/value", func(r chi.Router) {
				r.Route("/gauge", func(r chi.Router) {
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
				value, err := strconv.ParseFloat(string(body), 64)
				require.NoError(t, err)
				assert.Equal(t, tc.wantValue, value)
			}

			assert.Equal(t, tc.wantStatusCode, res.StatusCode)
		})
	}
}

func TestHandler_Update(t *testing.T) {

	tests := []struct {
		name           string
		givePath       string
		giveMethod     string
		wantStatusCode int
	}{
		{
			name:           "success",
			giveMethod:     http.MethodPost,
			givePath:       "/update/gauge/someMetric/1",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "method not allowed",
			giveMethod:     http.MethodGet,
			givePath:       "/update/gauge/someMetric/1",
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:           "method not found",
			giveMethod:     http.MethodPost,
			givePath:       "/update/gauge/1",
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "bad request",
			giveMethod:     http.MethodPost,
			givePath:       "/update/gauge/someMetric/metric",
			wantStatusCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()

			h := &Handler{
				Storage: &mockStorage{},
			}

			r.Route("/update", func(r chi.Router) {
				r.Route("/gauge", func(r chi.Router) {
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

			assert.Equal(t, res.StatusCode, tt.wantStatusCode)
		})
	}
}
