package value

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/vorotislav/alert-service/internal/http/handlers/value/mocks"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/vorotislav/alert-service/internal/http/middlewares"
	"github.com/vorotislav/alert-service/internal/model"
	"github.com/vorotislav/alert-service/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

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
