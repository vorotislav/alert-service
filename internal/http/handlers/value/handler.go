package value

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/vorotislav/alert-service/internal/model"
	"go.uber.org/zap"
	"net/http"
)

const (
	MetricCounter = "counter"
	MetricGauge   = "gauge"
)

type Storage interface {
	GetGaugeValue(name string) (float64, error)
	AllGaugeMetrics() ([]byte, error)
	GetCounterValue(name string) (int64, error)
	AllCounterMetrics() ([]byte, error)
}

type Handler struct {
	log     *zap.Logger
	storage Storage
}

func NewHandler(log *zap.Logger, storage Storage) *Handler {
	return &Handler{
		log:     log,
		storage: storage,
	}
}

func (h *Handler) Value(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.log.Info("Failed to get metrics",
			zap.Int("status code", http.StatusMethodNotAllowed),
			zap.Int("size", 0))

		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")

	var size int

	switch metricType {
	case MetricGauge:
		value, err := h.storage.GetGaugeValue(metricName)
		if err != nil {
			h.log.Info("Failed to update gauge metrics",
				zap.Int("status code", http.StatusNotFound),
				zap.Int("size", 0))

			http.Error(w, fmt.Sprintf("metrics %s if not found", metricName), http.StatusNotFound)

			return
		}

		size, err = w.Write([]byte(fmt.Sprintf("%v", value)))
		if err != nil {
			h.log.Info("Failed to get gauge metrics",
				zap.Int("status code", http.StatusBadRequest),
				zap.Int("size", 0))
		}

	case MetricCounter:
		value, err := h.storage.GetCounterValue(metricName)
		if err != nil {
			h.log.Info("Failed to get counter metrics",
				zap.Int("status code", http.StatusNotFound),
				zap.Int("size", 0))

			http.Error(w, fmt.Sprintf("metrics %s if not found", metricName), http.StatusNotFound)

			return
		}

		size, err = w.Write([]byte(fmt.Sprintf("%d", value)))
		if err != nil {
			h.log.Info("Failed to get counter metrics",
				zap.Int("status code", http.StatusInternalServerError),
				zap.Int("size", 0),
				zap.String("err", err.Error()))
		}
	default:
		h.log.Info("Failed get metrics: unknown metrics type")

		http.Error(w, "unknown metrics type", http.StatusBadRequest)
	}

	h.log.Info("Get gauge metrics",
		zap.Int("status code", http.StatusOK),
		zap.Int("size", size))

	w.WriteHeader(http.StatusOK)

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Add("Content-Type", "text/html")
	w.Header().Add("Content-Type", "charset=utf-8")
}

func (h *Handler) ValueJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.log.Info("Failed to get metrics",
			zap.Int("status code", http.StatusMethodNotAllowed),
			zap.Int("size", 0))

		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		h.log.Info("Failed to get metrics",
			zap.String("unknown Content-Type", contentType),
			zap.Int("status code", http.StatusBadRequest),
			zap.Int("size", 0))

		http.Error(w, "unknown Content-Type", http.StatusBadRequest)

		return
	}

	m := model.Metrics{}
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		h.log.Info("Failed to get metrics",
			zap.String("cannot decode metrics", err.Error()),
			zap.Int("status code", http.StatusBadRequest),
			zap.Int("size", 0))

		http.Error(w, "cannot decode metrics", http.StatusBadRequest)

		return
	}

	if m.ID == "" {
		h.log.Info("Failed to get metrics",
			zap.String("metrics ID is empty", ""),
			zap.Int("status code", http.StatusBadRequest),
			zap.Int("size", 0))

		http.Error(w, "metrics ID is empty", http.StatusBadRequest)

		return
	}

	switch m.MType {
	case MetricGauge:
		value, err := h.storage.GetGaugeValue(m.ID)
		if err != nil {
			h.log.Info("Failed to get gauge metrics",
				zap.Int("status code", http.StatusNotFound),
				zap.Int("size", 0))

			http.Error(w, fmt.Sprintf("metrics %s if not found", m.ID), http.StatusNotFound)

			return
		}

		m.Value = &value
	case MetricCounter:
		value, err := h.storage.GetCounterValue(m.ID)
		if err != nil {
			h.log.Info("Failed to get counter metrics",
				zap.Int("status code", http.StatusNotFound),
				zap.Int("size", 0))

			http.Error(w, fmt.Sprintf("metrics %s if not found", m.ID), http.StatusNotFound)

			return
		}

		m.Delta = &value
	default:
		h.log.Info("Failed to get metrics: unknown metrics type")

		http.Error(w, "unknown metrics type", http.StatusBadRequest)

		return
	}

	resp, err := json.Marshal(m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	size, err := w.Write(resp)
	if err != nil {
		h.log.Info("Error of write resp",
			zap.Error(err),
			zap.Int("status code", http.StatusInternalServerError),
			zap.Int("size", 0))
	}

	h.log.Info("Success get metrics",
		zap.Int("status code", http.StatusOK),
		zap.Int("size", size))
}

func (h *Handler) AllMetrics() ([]byte, error) {
	cntMetrics, err := h.storage.AllCounterMetrics()
	if err != nil {
		return nil, err
	}

	ggMetrics, err := h.storage.AllGaugeMetrics()
	if err != nil {
		return nil, err
	}

	resp := append(cntMetrics, ggMetrics...)

	return resp, nil
}
