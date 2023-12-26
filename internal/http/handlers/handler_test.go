package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/vorotislav/alert-service/internal/http/handlers/mocks"
	"github.com/vorotislav/alert-service/internal/http/middlewares"
	"github.com/vorotislav/alert-service/internal/model"
	"github.com/vorotislav/alert-service/internal/repository"
	srv "github.com/vorotislav/alert-service/internal/settings/server"
	"github.com/vorotislav/alert-service/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestHandler_Ping(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)

	tests := []struct {
		name           string
		prepareRepo    func(repository *mocks.MockRepository)
		wantStatusCode int
	}{
		{
			name: "success",
			prepareRepo: func(repository *mocks.MockRepository) {
				repository.EXPECT().Ping(gomock.Any()).Return(nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "error",
			prepareRepo: func(repository *mocks.MockRepository) {
				repository.EXPECT().Ping(gomock.Any()).Return(errors.New("some error"))
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

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

			r := chi.NewRouter()
			r.Use(middlewares.New(log))

			r.Route("/ping", func(r chi.Router) {
				r.Get("/", h.Ping)
			})

			server := httptest.NewServer(r)
			defer server.Close()

			request, err := http.NewRequest(http.MethodGet, server.URL+"/ping", http.NoBody)
			require.NoError(t, err)

			res, err := server.Client().Do(request)
			require.NoError(t, err)
			defer res.Body.Close()

			assert.Equal(t, tc.wantStatusCode, res.StatusCode)
		})
	}
}

func TestHandler_Updates(t *testing.T) {
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
			name: "success metrics",
			prepareRepo: func(repository *mocks.MockRepository) {
				repository.EXPECT().UpdateMetrics(gomock.Any(), gomock.Any()).Return(nil)
			},
			giveMethod:     http.MethodPost,
			giveBody:       []byte(`[{"id":"some counter", "mtype":"counter", "delta":1},{"id":"some gauge", "type":"gauge", "value":1.1}]`),
			wantStatusCode: http.StatusOK,
		},
		{
			name: "failed update",
			prepareRepo: func(repository *mocks.MockRepository) {
				repository.EXPECT().UpdateMetrics(gomock.Any(), gomock.Any()).Return(errors.New("some error"))
			},
			giveMethod:     http.MethodPost,
			giveBody:       []byte(`[{"id":"some counter", "mtype":"counter", "delta":1},{"id":"some counter", "type":"counter", "delta":1}]`),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "failed update cannot decode",
			giveMethod:     http.MethodPost,
			giveBody:       []byte(`[{"id":"some counter", "mtype":"counter", "delta":1},{"id":"some counter", "type":"counter", "delta":`),
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

			r.Route("/updates", func(r chi.Router) {
				r.Post("/", h.Updates)
			})

			server := httptest.NewServer(r)
			defer server.Close()

			request, err := http.NewRequest(tc.giveMethod, server.URL+"/updates", bytes.NewBuffer(tc.giveBody))
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")

			res, err := server.Client().Do(request)
			require.NoError(t, err)
			defer res.Body.Close()

			assert.Equal(t, tc.wantStatusCode, res.StatusCode)
		})
	}
}

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
				repository.EXPECT().UpdateMetric(gomock.Any(), gomock.Any()).Return(model.Metrics{ID: "someMetric"}, nil)
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
				repository.EXPECT().UpdateMetric(gomock.Any(), gomock.Any()).Return(model.Metrics{ID: "someMetric"}, nil)
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
				repository.EXPECT().UpdateMetric(gomock.Any(), gomock.Any()).Return(model.Metrics{ID: "some counter"}, nil)
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
				repository.EXPECT().UpdateMetric(gomock.Any(), gomock.Any()).Return(model.Metrics{ID: "some metrics"}, nil)
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

func TestHandler_Value(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)

	tests := []struct {
		name           string
		prepareRepo    func(repository *mocks.MockRepository)
		giveMethod     string
		givePath       string
		wantStatusCode int
		wantValue      float64
	}{
		{
			name: "success counter",
			prepareRepo: func(repository *mocks.MockRepository) {
				repository.EXPECT().GetCounterValue(gomock.Any(), gomock.Any()).Return(int64(15), nil)
			},
			giveMethod:     http.MethodGet,
			givePath:       "/value/counter/PollCount",
			wantStatusCode: http.StatusOK,
			wantValue:      15,
		},
		{
			name:           "not allowed counter",
			giveMethod:     http.MethodPost,
			givePath:       "/value/counter/PollCount",
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name: "success",
			prepareRepo: func(repository *mocks.MockRepository) {
				repository.EXPECT().GetGaugeValue(gomock.Any(), gomock.Any()).Return(float64(11.1), nil)
			},
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

			r.Route("/value", func(r chi.Router) {
				r.Route("/{metricType}", func(r chi.Router) {
					r.Get("/{metricName}", h.Value)
				})

			})

			server := httptest.NewServer(r)
			defer server.Close()

			request, err := http.NewRequest(tc.giveMethod, server.URL+tc.givePath, http.NoBody)
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

func TestHandler_ValueJSON(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)

	getSuccessCounterValue := func() *int64 {
		value := int64(15)
		return &value
	}

	getSuccessGaugeValue := func() *float64 {
		value := float64(11.1)

		return &value
	}

	tests := []struct {
		name           string
		prepareRepo    func(repository *mocks.MockRepository)
		giveMethod     string
		giveBody       []byte
		wantStatusCode int
		wantMetric     model.Metrics
	}{
		{
			name: "success counter",
			prepareRepo: func(repository *mocks.MockRepository) {
				repository.EXPECT().GetCounterValue(gomock.Any(), gomock.Any()).Return(int64(15), nil)
			},
			giveMethod:     http.MethodPost,
			giveBody:       []byte(`{"id":"PollCount", "type":"counter"}`),
			wantStatusCode: http.StatusOK,
			wantMetric: model.Metrics{
				ID:    "PollCount",
				MType: "counter",
				Delta: getSuccessCounterValue(),
			},
		},
		{
			name:           "not allowed",
			giveMethod:     http.MethodPut,
			giveBody:       []byte(`{"id":"PollCount", "type":"counter"}`),
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name: "Counter not found",
			prepareRepo: func(repository *mocks.MockRepository) {
				repository.EXPECT().GetCounterValue(gomock.Any(), gomock.Any()).Return(int64(0), errors.New("some error"))
			},
			giveMethod:     http.MethodPost,
			giveBody:       []byte(`{"id":"some name", "type":"counter"}`),
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "success gauge",
			prepareRepo: func(repository *mocks.MockRepository) {
				repository.EXPECT().GetGaugeValue(gomock.Any(), gomock.Any()).Return(float64(11.1), nil)
			},
			giveBody:       []byte(`{"id":"mymetric", "type":"gauge"}`),
			giveMethod:     http.MethodPost,
			wantStatusCode: http.StatusOK,
			wantMetric: model.Metrics{
				ID:    "mymetric",
				MType: "gauge",
				Value: getSuccessGaugeValue(),
			},
		},
		{
			name:           "bad request of metrics type",
			giveMethod:     http.MethodPost,
			giveBody:       []byte(`{"id":"name", "type":"azaza"}`),
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
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
			r.Use(middlewares.CompressMiddleware)

			r.Route("/value", func(r chi.Router) {
				r.Post("/", h.ValueJSON)
			})

			server := httptest.NewServer(r)
			defer server.Close()

			request, err := http.NewRequest(tc.giveMethod, server.URL+"/value", bytes.NewBuffer(tc.giveBody))
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")

			res, err := server.Client().Do(request)
			require.NoError(t, err)
			defer res.Body.Close()
			assert.Equal(t, tc.wantStatusCode, res.StatusCode)
			if res.StatusCode == http.StatusOK {
				m := model.Metrics{}

				bodyRaw, err := io.ReadAll(res.Body)
				require.NoError(t, err)

				if strings.Contains(res.Header.Get("Content-Encoding"), "gzip") {
					bodyRaw, err = utils.Decompress(bodyRaw)
					require.NoError(t, err)
				}

				err = json.Unmarshal(bodyRaw, &m)
				require.NoError(t, err)
				assert.Equal(t, tc.wantMetric, m)
			}
		})
	}
}

func TestNewHandler(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockRepository(ctrl)

	h := NewHandler(log, m)
	require.NotNil(t, h)
}

func BenchmarkHandler_Ping(b *testing.B) {
	interval := 0
	restore := false
	log, _ := zap.NewDevelopment()
	rep, _ := repository.NewRepository(context.Background(), log, &srv.Settings{
		Address:         "",
		StoreInterval:   &interval,
		FileStoragePath: "",
		Restore:         &restore,
		DatabaseDSN:     "",
		HashKey:         "",
	})

	h := &Handler{
		repo: rep,
	}

	r := chi.NewRouter()

	r.Route("/ping", func(r chi.Router) {
		r.Get("/", h.Ping)
	})

	server := httptest.NewServer(r)
	defer server.Close()

	for i := 0; i < b.N; i++ {
		request, _ := http.NewRequest(http.MethodGet, server.URL+"/ping", http.NoBody)

		res, err := server.Client().Do(request)
		if err != nil {
			defer res.Body.Close()
		}

	}
}

func BenchmarkHandler_Updates(b *testing.B) {
	interval := 0
	restore := false
	body := []byte(`[{"id":"some counter", "mtype":"counter", "delta":1},{"id":"some gauge", "type":"gauge", "value":1.1}]`)
	log, _ := zap.NewDevelopment()
	rep, _ := repository.NewRepository(context.Background(), log, &srv.Settings{
		Address:         "",
		StoreInterval:   &interval,
		FileStoragePath: "",
		Restore:         &restore,
		DatabaseDSN:     "",
		HashKey:         "",
	})

	h := &Handler{
		repo: rep,
		log:  log,
	}

	r := chi.NewRouter()

	r.Route("/updates", func(r chi.Router) {
		r.Post("/", h.Updates)
	})

	server := httptest.NewServer(r)
	defer server.Close()

	for i := 0; i < b.N; i++ {
		request, _ := http.NewRequest(http.MethodPost, server.URL+"/updates", bytes.NewBuffer(body))
		request.Header.Set("Content-Type", "application/json")

		res, err := server.Client().Do(request)
		if err != nil {
			defer res.Body.Close()
		}

	}
}

func BenchmarkHandler_Update(b *testing.B) {
	interval := 0
	restore := false
	log, _ := zap.NewDevelopment()
	rep, _ := repository.NewRepository(context.Background(), log, &srv.Settings{
		Address:         "",
		StoreInterval:   &interval,
		FileStoragePath: "",
		Restore:         &restore,
		DatabaseDSN:     "",
		HashKey:         "",
	})

	h := &Handler{
		repo: rep,
		log:  log,
	}

	r := chi.NewRouter()

	r.Route("/update", func(r chi.Router) {
		r.Route("/{metricType}", func(r chi.Router) {
			r.Route("/{metricName}", func(r chi.Router) {
				r.Post("/{metricValue}", h.Update)
			})
		})
	})

	server := httptest.NewServer(r)
	defer server.Close()

	for i := 0; i < b.N; i++ {
		request, _ := http.NewRequest(http.MethodPost, server.URL+"/update/counter/1", http.NoBody)
		request.Header.Set("Content-Type", "application/json")

		res, err := server.Client().Do(request)
		if err != nil {
			defer res.Body.Close()
		}

	}
}

func BenchmarkHandler_UpdateJSON(b *testing.B) {
	interval := 0
	restore := false
	body := []byte(`{"id":"some counter", "type":"counter", "delta":1}`)
	log, _ := zap.NewDevelopment()
	rep, _ := repository.NewRepository(context.Background(), log, &srv.Settings{
		Address:         "",
		StoreInterval:   &interval,
		FileStoragePath: "",
		Restore:         &restore,
		DatabaseDSN:     "",
		HashKey:         "",
	})

	h := &Handler{
		repo: rep,
		log:  log,
	}

	r := chi.NewRouter()

	r.Route("/updates", func(r chi.Router) {
		r.Post("/", h.Updates)
	})

	server := httptest.NewServer(r)
	defer server.Close()

	for i := 0; i < b.N; i++ {
		request, _ := http.NewRequest(http.MethodPost, server.URL+"/update", bytes.NewBuffer(body))
		request.Header.Set("Content-Type", "application/json")

		res, err := server.Client().Do(request)
		if err != nil {
			defer res.Body.Close()
		}

	}
}
