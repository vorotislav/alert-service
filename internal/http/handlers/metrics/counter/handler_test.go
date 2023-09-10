package counter

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
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

func TestHandler_Update(t *testing.T) {
	tests := []struct {
		name           string
		giveRequest    *http.Request
		wantStatusCode int
	}{
		{
			name:           "success",
			giveRequest:    httptest.NewRequest(http.MethodPost, "/update/counter/someMetric/1", nil),
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "method not allowed",
			giveRequest:    httptest.NewRequest(http.MethodGet, "/update/counter/someMetric/1", nil),
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:           "method not found",
			giveRequest:    httptest.NewRequest(http.MethodPost, "/update/counter/1", nil),
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "bad request",
			giveRequest:    httptest.NewRequest(http.MethodPost, "/update/counter/someMetric/metric", nil),
			wantStatusCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			h := &Handler{
				s: &mockStorage{},
			}
			h.Update(w, tt.giveRequest)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, res.StatusCode, tt.wantStatusCode)
		})
	}
}
