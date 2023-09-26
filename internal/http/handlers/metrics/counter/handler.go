package counter

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type Storage interface {
	UpdateCounter(name string, value int64) error
	GetCounterValue(name string) (int64, error)
	AllCounterMetrics() ([]byte, error)
}

type Handler struct {
	s Storage
}

func NewHandler(storage Storage) *Handler {
	return &Handler{
		s: storage,
	}
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	metricName := chi.URLParam(r, "metricName")
	metricValue, err := strconv.ParseInt(chi.URLParam(r, "metricValue"), 10, 64)

	if err != nil {
		http.Error(w, fmt.Errorf("cannot convert metric value: %w", err).Error(), http.StatusBadRequest)

		return
	}

	if err := h.s.UpdateCounter(metricName, metricValue); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Add("Content-Type", "charset=utf-8")
}

func (h *Handler) Value(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	metricName := chi.URLParam(r, "metricName")
	value, err := h.s.GetCounterValue(metricName)
	if err != nil {
		http.Error(w, fmt.Sprintf("metric %s if not found", metricName), http.StatusNotFound)

		return
	}

	w.Write([]byte(fmt.Sprintf("%d", value)))
	w.WriteHeader(http.StatusOK)

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Add("Content-Type", "charset=utf-8")
}

func (h *Handler) AllMetrics() ([]byte, error) {
	return h.s.AllCounterMetrics()
}
