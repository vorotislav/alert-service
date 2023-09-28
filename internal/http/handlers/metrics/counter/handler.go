package counter

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type Storage interface {
	UpdateCounter(name string, value int64) error
	GetCounterValue(name string) (int64, error)
	AllCounterMetrics() ([]byte, error)
}

type Handler struct {
	log *zap.Logger
	s   Storage
}

func NewHandler(log *zap.Logger, storage Storage) *Handler {
	return &Handler{
		log: log,
		s:   storage,
	}
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.log.Info("Failed to update counter metric",
			zap.Int("status code", http.StatusMethodNotAllowed),
			zap.Int("size", 0))

		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	metricName := chi.URLParam(r, "metricName")
	metricValue, err := strconv.ParseInt(chi.URLParam(r, "metricValue"), 10, 64)

	if err != nil {
		h.log.Info("Failed to update counter metric",
			zap.Int("status code", http.StatusBadRequest),
			zap.Int("size", 0))

		http.Error(w, fmt.Errorf("cannot convert metric value: %w", err).Error(), http.StatusBadRequest)

		return
	}

	if err := h.s.UpdateCounter(metricName, metricValue); err != nil {
		h.log.Info("Failed to update counter metric",
			zap.Int("status code", http.StatusInternalServerError),
			zap.Int("size", 0))

		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	h.log.Info("Success update counter metric",
		zap.Int("status code", http.StatusOK),
		zap.Int("size", 0))

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Add("Content-Type", "charset=utf-8")
}

func (h *Handler) Value(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.log.Info("Failed to get counter metric",
			zap.Int("status code", http.StatusMethodNotAllowed),
			zap.Int("size", 0))

		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	metricName := chi.URLParam(r, "metricName")
	value, err := h.s.GetCounterValue(metricName)
	if err != nil {
		h.log.Info("Failed to get counter metric",
			zap.Int("status code", http.StatusNotFound),
			zap.Int("size", 0))

		http.Error(w, fmt.Sprintf("metric %s if not found", metricName), http.StatusNotFound)

		return
	}

	size, err := w.Write([]byte(fmt.Sprintf("%d", value)))
	if err != nil {
		h.log.Info("Failed to get counter metric",
			zap.Int("status code", http.StatusInternalServerError),
			zap.Int("size", 0),
			zap.String("err", err.Error()))
	}

	h.log.Info("Success get counter metric",
		zap.Int("status code", http.StatusOK),
		zap.Int("size", size))

	w.WriteHeader(http.StatusOK)

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Add("Content-Type", "charset=utf-8")
}

func (h *Handler) AllMetrics() ([]byte, error) {
	return h.s.AllCounterMetrics()
}
