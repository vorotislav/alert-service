package update

import (
	"bytes"
	"github.com/golang/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vorotislav/alert-service/internal/http/handlers/update/mocks"
	"github.com/vorotislav/alert-service/internal/http/middlewares"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestHandler_Update(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)

	cases := []struct {
		name           string
		prepareRepo    func(repository *mocks.MockRepository)
		giveMethod     string
		givePath       string
		wantStatusCode int
	}{
		{
			name: "success counter",
			prepareRepo: func(repository *mocks.MockRepository) {
				repository.EXPECT().UpdateCounter(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			},
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
			givePath:       "/update/counter/someMetric/metrics",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "success gauge",
			prepareRepo: func(repository *mocks.MockRepository) {
				repository.EXPECT().UpdateGauge(gomock.Any(), gomock.Any(), gomock.Any()).Return(float64(1), nil)
			},
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
			givePath:       "/update/gauge/someMetric/metrics",
			wantStatusCode: http.StatusBadRequest,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := chi.NewRouter()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := mocks.NewMockRepository(ctrl)
			if tc.prepareRepo != nil {
				tc.prepareRepo(m)
			}

			h := &Handler{
				log:  log,
				repo: m,
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

			request, err := http.NewRequest(tc.giveMethod, server.URL+tc.givePath, nil)
			require.NoError(t, err)

			res, err := server.Client().Do(request)
			require.NoError(t, err)
			defer res.Body.Close()

			assert.Equal(t, tc.wantStatusCode, res.StatusCode)
		})
	}
}

func TestHandler_UpdateJSON(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)

	cases := []struct {
		name           string
		prepareRepo    func(repository *mocks.MockRepository)
		giveMethod     string
		giveBody       []byte
		wantStatusCode int
	}{
		{
			name: "success counter",
			prepareRepo: func(repository *mocks.MockRepository) {
				repository.EXPECT().UpdateCounter(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			},
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
			name: "success gauge",
			prepareRepo: func(repository *mocks.MockRepository) {
				repository.EXPECT().UpdateGauge(gomock.Any(), gomock.Any(), gomock.Any()).Return(float64(1), nil)
			},
			giveMethod:     http.MethodPost,
			giveBody:       []byte(`{"id":"some metrics", "type":"gauge", "value":1}`),
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "method not allowed gauge",
			giveMethod:     http.MethodGet,
			giveBody:       []byte(`{"id":"some metrics", "type":"gauge", "value":1}`),
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:           "bad request gauge",
			giveMethod:     http.MethodPost,
			giveBody:       []byte(`{"id":"some metrics", "type":"gauge"}`),
			wantStatusCode: http.StatusBadRequest,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := chi.NewRouter()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := mocks.NewMockRepository(ctrl)
			if tc.prepareRepo != nil {
				tc.prepareRepo(m)
			}

			h := &Handler{
				log:  log,
				repo: m,
			}

			r.Use(middlewares.New(log))

			r.Route("/update", func(r chi.Router) {
				r.Post("/", h.UpdateJSON)
			})

			server := httptest.NewServer(r)
			defer server.Close()

			request, err := http.NewRequest(tc.giveMethod, server.URL+"/update", bytes.NewBuffer(tc.giveBody))
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")

			res, err := server.Client().Do(request)
			require.NoError(t, err)
			defer res.Body.Close()

			assert.Equal(t, tc.wantStatusCode, res.StatusCode)
		})
	}
}
