package update

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/vorotislav/alert-service/internal/model"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

const (
	MetricCounter = "counter"
	MetricGauge   = "gauge"
)

type Storage interface {
	UpdateGauge(name string, value float64) (float64, error)
	UpdateCounter(name string, value int64) (int64, error)
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

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")

	switch metricType {
	case MetricGauge:
		h.updateGauge(w, r)
	case MetricCounter:
		h.updateCounter(w, r)
	default:
		h.log.Info("Failed update metrics: unknown metrics type")

		http.Error(w, "unknown metrics type", http.StatusBadRequest)
		return
	}
}

func (h *Handler) updateCounter(w http.ResponseWriter, r *http.Request) {
	metricName := chi.URLParam(r, "metricName")
	metricValue, err := strconv.ParseInt(chi.URLParam(r, "metricValue"), 10, 64)

	if err != nil {
		h.log.Info("Failed to update counter metrics",
			zap.Int("status code", http.StatusBadRequest),
			zap.Int("size", 0))

		http.Error(w, fmt.Errorf("cannot convert metrics value: %w", err).Error(), http.StatusBadRequest)

		return
	}

	_, err = h.storage.UpdateCounter(metricName, metricValue)
	if err != nil {
		h.log.Info("Failed to update counter metrics",
			zap.Int("status code", http.StatusInternalServerError),
			zap.Int("size", 0))

		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	h.log.Info("Success update counter metrics",
		zap.Int("status code", http.StatusOK),
		zap.Int("size", 0))

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Add("Content-Type", "charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) updateGauge(w http.ResponseWriter, r *http.Request) {
	metricName := chi.URLParam(r, "metricName")
	metricValue, err := strconv.ParseFloat(chi.URLParam(r, "metricValue"), 64)

	if err != nil {
		h.log.Info("Failed to update gauge metrics",
			zap.Int("status code", http.StatusBadRequest),
			zap.Int("size", 0))

		http.Error(w, fmt.Errorf("cannot convert metrics value: %w", err).Error(), http.StatusBadRequest)

		return
	}

	_, err = h.storage.UpdateGauge(metricName, metricValue)
	if err != nil {
		h.log.Info("Failed to update gauge metrics",
			zap.Int("status code", http.StatusInternalServerError),
			zap.Int("size", 0))

		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Add("Content-Type", "charset=utf-8")

	w.WriteHeader(http.StatusOK)

	h.log.Info("Success update gauge metrics",
		zap.Int("status code", http.StatusOK),
		zap.Int("size", 0))
}

func (h *Handler) UpdateJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.log.Info("Failed to update metrics",
			zap.Int("status code", http.StatusMethodNotAllowed),
			zap.Int("size", 0))

		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		h.log.Info("Failed to update metrics",
			zap.String("unknown Content-Type", contentType),
			zap.Int("status code", http.StatusBadRequest),
			zap.Int("size", 0))

		http.Error(w, "unknown Content-Type", http.StatusBadRequest)

		return
	}

	m := model.Metrics{}
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		h.log.Info("Failed to update metrics",
			zap.String("cannot decode metrics", err.Error()),
			zap.Int("status code", http.StatusBadRequest),
			zap.Int("size", 0))

		http.Error(w, "cannot decode metrics", http.StatusBadRequest)

		return
	}

	switch m.MType {
	case MetricGauge:
		if m.Value == nil {
			h.log.Info("Failed update metrics: no metrics value")

			http.Error(w, "no metrics value", http.StatusBadRequest)

			return
		}

		newValue, err := h.storage.UpdateGauge(m.ID, *m.Value)
		if err != nil {
			h.log.Info("Failed update metrics",
				zap.Error(err))

			http.Error(w, fmt.Sprintf("no metrics value: %s", err.Error()), http.StatusBadRequest)

			return
		}

		m.Value = &newValue

	case MetricCounter:
		if m.Delta == nil {
			h.log.Info("Failed update metrics: no metrics value")

			http.Error(w, "no metrics value", http.StatusBadRequest)

			return
		}

		newValue, err := h.storage.UpdateCounter(m.ID, *m.Delta)
		if err != nil {
			h.log.Info("Failed update metrics",
				zap.Error(err))

			http.Error(w, fmt.Sprintf("no metrics value: %s", err.Error()), http.StatusBadRequest)

			return
		}

		m.Delta = &newValue

	default:
		h.log.Info("Failed update metrics: unknown metrics type")

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

	h.log.Info("Success update metrics",
		zap.Int("status code", http.StatusOK),
		zap.Int("size", size))
}
