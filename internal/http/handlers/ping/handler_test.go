package ping

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vorotislav/alert-service/internal/http/handlers/ping/mocks"
	"github.com/vorotislav/alert-service/internal/http/middlewares"
	"net/http"
	"net/http/httptest"
	"testing"

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
