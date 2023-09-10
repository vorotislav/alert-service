package gauge

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockStorage struct {
}

func (m *mockStorage) UpdateGauge(name string, value float64) error {
	return nil
}

func (m *mockStorage) GetGaugeValue(name string) (float64, error) {
	return 0, nil
}

func (m *mockStorage) AllGaugeMetrics() ([]byte, error) {
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
			giveRequest:    httptest.NewRequest(http.MethodPost, "/update/gauge/someMetric/1", nil),
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "method not allowed",
			giveRequest:    httptest.NewRequest(http.MethodGet, "/update/gauge/someMetric/1", nil),
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:           "method not found",
			giveRequest:    httptest.NewRequest(http.MethodPost, "/update/gauge/1", nil),
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "bad request",
			giveRequest:    httptest.NewRequest(http.MethodPost, "/update/gauge/someMetric/metric", nil),
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
