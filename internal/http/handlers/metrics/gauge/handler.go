package gauge

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type Storage interface {
	UpdateGauge(name string, value float64) error
	GetGaugeValue(name string) (float64, error)
	AllGaugeMetrics() ([]byte, error)
}

type Handler struct {
	log     *zap.Logger
	Storage Storage
}

func NewHandler(log *zap.Logger, storage Storage) *Handler {
	return &Handler{
		log:     log,
		Storage: storage,
	}
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.log.Info("Failed to update gauge metric",
			zap.Int("status code", http.StatusMethodNotAllowed),
			zap.Int("size", 0))

		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	metricName := chi.URLParam(r, "metricName")
	metricValue, err := strconv.ParseFloat(chi.URLParam(r, "metricValue"), 64)

	if err != nil {
		h.log.Info("Failed to update gauge metric",
			zap.Int("status code", http.StatusBadRequest),
			zap.Int("size", 0))

		http.Error(w, fmt.Errorf("cannot convert metric value: %w", err).Error(), http.StatusBadRequest)

		return
	}

	if err := h.Storage.UpdateGauge(metricName, metricValue); err != nil {
		h.log.Info("Failed to update gauge metric",
			zap.Int("status code", http.StatusInternalServerError),
			zap.Int("size", 0))

		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Add("Content-Type", "charset=utf-8")

	w.WriteHeader(http.StatusOK)

	h.log.Info("Success update gauge metric",
		zap.Int("status code", http.StatusOK),
		zap.Int("size", 0))
}

func (h *Handler) Value(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.log.Info("Failed to get gauge metric",
			zap.Int("status code", http.StatusMethodNotAllowed),
			zap.Int("size", 0))

		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	metricName := chi.URLParam(r, "metricName")
	value, err := h.Storage.GetGaugeValue(metricName)
	if err != nil {
		h.log.Info("Failed to update gauge metric",
			zap.Int("status code", http.StatusNotFound),
			zap.Int("size", 0))

		http.Error(w, fmt.Sprintf("metric %s if not found", metricName), http.StatusNotFound)

		return
	}

	size, err := w.Write([]byte(fmt.Sprintf("%v", value)))
	if err != nil {
		h.log.Info("Failed to get gauge metric",
			zap.Int("status code", http.StatusBadRequest),
			zap.Int("size", 0))
	}

	h.log.Info("Get gauge metric",
		zap.Int("status code", http.StatusOK),
		zap.Int("size", size))

	w.WriteHeader(http.StatusOK)

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Add("Content-Type", "charset=utf-8")
}

func (h *Handler) AllMetrics() ([]byte, error) {
	return h.Storage.AllGaugeMetrics()
}
