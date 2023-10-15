package updates

import (
	"bytes"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vorotislav/alert-service/internal/http/handlers/updates/mocks"
	"github.com/vorotislav/alert-service/internal/http/middlewares"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
			giveBody:       []byte(`[{"id":"some counter", "type":"counter", "delta":1},{"id":"some counter", "type":"counter", "delta":1}]`),
			wantStatusCode: http.StatusOK,
		},
		{
			name: "failed update",
			prepareRepo: func(repository *mocks.MockRepository) {
				repository.EXPECT().UpdateMetrics(gomock.Any(), gomock.Any()).Return(errors.New("some error"))
			},
			giveMethod:     http.MethodPost,
			giveBody:       []byte(`[{"id":"some counter", "type":"counter", "delta":1},{"id":"some counter", "type":"counter", "delta":1}]`),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "failed update cannot decode",
			giveMethod:     http.MethodPost,
			giveBody:       []byte(`[{"id":"some counter", "type":"counter", "delta":1},{"id":"some counter", "type":"counter", "delta":`),
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
