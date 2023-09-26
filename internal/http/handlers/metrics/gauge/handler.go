package gauge

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type Storage interface {
	UpdateGauge(name string, value float64) error
	GetGaugeValue(name string) (float64, error)
	AllGaugeMetrics() ([]byte, error)
}

type Handler struct {
	Storage Storage
}

func NewHandler(storage Storage) *Handler {
	return &Handler{
		Storage: storage,
	}
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	fmt.Println(chi.URLParam(r, "metricName"), chi.URLParam(r, "metricValue"))

	metricName := chi.URLParam(r, "metricName")
	metricValue, err := strconv.ParseFloat(chi.URLParam(r, "metricValue"), 64)

	if err != nil {
		http.Error(w, fmt.Errorf("cannot convert metric value: %w", err).Error(), http.StatusBadRequest)

		return
	}

	if err := h.Storage.UpdateGauge(metricName, metricValue); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Add("Content-Type", "charset=utf-8")

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Value(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	metricName := chi.URLParam(r, "metricName")
	value, err := h.Storage.GetGaugeValue(metricName)
	if err != nil {
		http.Error(w, fmt.Sprintf("metric %s if not found", metricName), http.StatusNotFound)

		return
	}

	w.Write([]byte(fmt.Sprintf("%v", value)))
	w.WriteHeader(http.StatusOK)

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Add("Content-Type", "charset=utf-8")
}

func (h *Handler) AllMetrics() ([]byte, error) {
	return h.Storage.AllGaugeMetrics()
}
