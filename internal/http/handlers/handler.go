package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
	"time"

	"github.com/vorotislav/alert-service/internal/model"

	"go.uber.org/zap"
)

const (
	MetricCounter = "counter"
	MetricGauge   = "gauge"
)

const (
	queryRepoTimeout = time.Second + 2
)

type Repository interface {
	UpdateMetric(ctx context.Context, metric model.Metrics) (model.Metrics, error)
	GetCounterValue(ctx context.Context, name string) (int64, error)
	GetGaugeValue(ctx context.Context, name string) (float64, error)
	AllMetrics(ctx context.Context) ([]byte, error)
	Ping(ctx context.Context) error
	UpdateMetrics(ctx context.Context, metrics []model.Metrics) error
}

type Handler struct {
	log  *zap.Logger
	repo Repository
}

func NewHandler(log *zap.Logger, r Repository) *Handler {
	return &Handler{
		log:  log,
		repo: r,
	}
}

func (h *Handler) logInfo(msg string, status, size int) {
	h.log.Info(msg, zap.Int("status code", status), zap.Int("size", size))
}

func setContentType(w http.ResponseWriter, contentType string) {
	w.Header().Set("Content-Type", contentType)
	if contentType == "text/plain" {
		w.Header().Add("Content-Type", "charset=utf-8")
	}
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), queryRepoTimeout)
	defer cancel()
	err := h.repo.Ping(ctx)
	if err != nil {
		http.Error(w, "repository is not available", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")

	switch metricType {
	case MetricGauge:
		h.updateGauge(w, r)
	case MetricCounter:
		h.updateCounter(w, r)
	default:
		h.logInfo("Failed update metrics: unknown metrics type", http.StatusBadRequest, 0)

		http.Error(w, "unknown metrics type", http.StatusBadRequest)
		return
	}
}

func (h *Handler) updateCounter(w http.ResponseWriter, r *http.Request) {
	metricName := chi.URLParam(r, "metricName")
	metricValue, err := strconv.ParseInt(chi.URLParam(r, "metricValue"), 10, 64)

	if err != nil {
		h.logInfo("Failed to update counter metrics", http.StatusBadRequest, 0)

		http.Error(w, fmt.Sprintf("cannot convert metrics value: %s", err.Error()), http.StatusBadRequest)

		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), queryRepoTimeout)
	defer cancel()
	_, err = h.repo.UpdateMetric(ctx, model.Metrics{
		ID:    metricName,
		MType: model.MetricCounter,
		Delta: &metricValue,
	})
	if err != nil {
		h.logInfo("Failed to update counter metrics", http.StatusInternalServerError, 0)

		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	h.logInfo("Success update counter metrics", http.StatusOK, 0)

	setContentType(w, "text/plain")
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

	ctx, cancel := context.WithTimeout(r.Context(), queryRepoTimeout)
	defer cancel()
	_, err = h.repo.UpdateMetric(ctx, model.Metrics{
		ID:    metricName,
		MType: model.MetricGauge,
		Value: &metricValue,
	})
	if err != nil {
		h.logInfo("Failed to update gauge metrics", http.StatusInternalServerError, 0)

		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	setContentType(w, "text/plain")

	w.WriteHeader(http.StatusOK)

	h.logInfo("Success update gauge metrics", http.StatusOK, 0)
}

func (h *Handler) UpdateJSON(w http.ResponseWriter, r *http.Request) {
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		h.logInfo(fmt.Sprintf("Failed to update metrics: unknown ContentType %s", contentType),
			http.StatusBadRequest, 0)

		http.Error(w, "unknown Content-Type", http.StatusBadRequest)

		return
	}

	m := model.Metrics{}
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		h.logInfo(fmt.Sprintf("Failed to update metrics: cannot decode metrics: %s", err.Error()),
			http.StatusBadRequest, 0)

		http.Error(w, "cannot decode metrics", http.StatusBadRequest)

		return
	}

	var err error

	ctx, cancel := context.WithTimeout(r.Context(), queryRepoTimeout)
	defer cancel()

	switch m.MType {
	case MetricGauge:
		if m.Value == nil {
			h.logInfo("Failed update metrics: no metrics value", http.StatusBadRequest, 0)

			http.Error(w, "no metrics value", http.StatusBadRequest)

			return
		}

		m, err = h.repo.UpdateMetric(ctx, m)
		if err != nil {
			h.logInfo(fmt.Sprintf("Failed update metrics: %s", err.Error()), http.StatusBadRequest, 0)

			http.Error(w, fmt.Sprintf("update metrics value: %s", err.Error()), http.StatusBadRequest)

			return
		}

	case MetricCounter:
		if m.Delta == nil {
			h.logInfo("Failed update metrics: no metrics value", http.StatusBadRequest, 0)

			http.Error(w, "no metrics value", http.StatusBadRequest)

			return
		}

		m, err = h.repo.UpdateMetric(ctx, m)
		if err != nil {
			h.logInfo(fmt.Sprintf("Failed update metrics: %s", err.Error()), http.StatusBadRequest, 0)

			http.Error(w, fmt.Sprintf("update metrics value: %s", err.Error()), http.StatusBadRequest)

			return
		}

	default:
		h.logInfo("Failed update metrics: unknown metrics type", http.StatusBadRequest, 0)

		http.Error(w, "unknown metrics type", http.StatusBadRequest)

		return
	}

	resp, err := json.Marshal(m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	setContentType(w, "application/json")
	w.WriteHeader(http.StatusOK)
	size, err := w.Write(resp)
	if err != nil {
		h.logInfo(fmt.Sprintf("Error of write resp: %s", err.Error()), http.StatusInternalServerError, 0)
	}

	h.logInfo("Success update metrics", http.StatusOK, size)
}

func (h *Handler) Updates(w http.ResponseWriter, r *http.Request) {
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		h.logInfo(fmt.Sprintf("Failed to update metrics: unknown Content-Type: %s", contentType),
			http.StatusBadRequest, 0)

		http.Error(w, "unknown Content-Type", http.StatusBadRequest)

		return
	}

	metrics := make([]model.Metrics, 0)
	err := json.NewDecoder(r.Body).Decode(&metrics)
	if err != nil {
		h.logInfo(fmt.Sprintf("Failed to update metrics: cannot decode metrics: %s", err.Error()),
			http.StatusBadRequest, 0)

		http.Error(w, "cannot decode body", http.StatusBadRequest)

		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), queryRepoTimeout)
	defer cancel()
	if err := h.repo.UpdateMetrics(ctx, metrics); err != nil {
		h.logInfo(fmt.Sprintf("Failed to update metrics: %s", err.Error()),
			http.StatusBadRequest, 0)

		http.Error(w, fmt.Sprintf("cannot update metrics: %s", err.Error()), http.StatusBadRequest)

		return
	}

	setContentType(w, "application/json")
	w.WriteHeader(http.StatusOK)
	size, err := w.Write([]byte(`{}`))
	if err != nil {
		h.logInfo(fmt.Sprintf("Error of write resp: %s", err.Error()),
			http.StatusInternalServerError, 0)
	}

	h.logInfo("Success update metrics", http.StatusOK, size)
}

func (h *Handler) Value(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")

	var size int

	ctx, cancel := context.WithTimeout(r.Context(), queryRepoTimeout)
	defer cancel()

	switch metricType {
	case MetricGauge:
		value, err := h.repo.GetGaugeValue(ctx, metricName)
		if err != nil {
			h.logInfo("Failed to update gauge metrics", http.StatusNotFound, 0)

			http.Error(w, fmt.Sprintf("metrics %s if not found", metricName), http.StatusNotFound)

			return
		}

		size, err = w.Write([]byte(fmt.Sprintf("%v", value)))
		if err != nil {
			h.logInfo(fmt.Sprintf("Failed to get gauge metrics: %s", err.Error()),
				http.StatusBadRequest, 0)
		}

	case MetricCounter:
		value, err := h.repo.GetCounterValue(ctx, metricName)
		if err != nil {
			h.logInfo(fmt.Sprintf("Failed to get counter metrics: %s", err.Error()),
				http.StatusNotFound, 0)

			http.Error(w, fmt.Sprintf("metrics %s if not found", metricName), http.StatusNotFound)

			return
		}

		size, err = w.Write([]byte(fmt.Sprintf("%d", value)))
		if err != nil {
			h.logInfo(fmt.Sprintf("Failed to get counter metrics: %s", err.Error()),
				http.StatusInternalServerError, 0)
		}
	default:
		h.logInfo("Failed get metrics: unknown metrics type", http.StatusBadRequest, 0)

		http.Error(w, "unknown metrics type", http.StatusBadRequest)
	}

	h.logInfo("Get gauge metrics", http.StatusOK, size)

	setContentType(w, "text/plain")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ValueJSON(w http.ResponseWriter, r *http.Request) {
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		h.logInfo(fmt.Sprintf("Failed to get metrics: unknown Content-Type: %s", contentType),
			http.StatusBadRequest, 0)

		http.Error(w, "unknown Content-Type", http.StatusBadRequest)

		return
	}

	m := model.Metrics{}
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		h.logInfo(fmt.Sprintf("Failed to get metrics: cannot decode metrics: %s", err.Error()),
			http.StatusBadRequest, 0)

		http.Error(w, "cannot decode metrics", http.StatusBadRequest)

		return
	}

	if m.ID == "" {
		h.logInfo("Failed to get metrics: metrics ID is empty", http.StatusBadRequest, 0)

		http.Error(w, "metrics ID is empty", http.StatusBadRequest)

		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), queryRepoTimeout)
	defer cancel()

	switch m.MType {
	case MetricGauge:
		value, err := h.repo.GetGaugeValue(ctx, m.ID)
		if err != nil {
			h.logInfo("Failed to get gauge metrics", http.StatusNotFound, 0)

			http.Error(w, fmt.Sprintf("metrics %s if not found", m.ID), http.StatusNotFound)

			return
		}

		m.Value = &value
	case MetricCounter:
		value, err := h.repo.GetCounterValue(ctx, m.ID)
		if err != nil {
			h.logInfo("Failed to get counter metrics", http.StatusNotFound, 0)

			http.Error(w, fmt.Sprintf("metrics %s if not found", m.ID), http.StatusNotFound)

			return
		}

		m.Delta = &value
	default:
		h.logInfo("Failed to get metrics: unknown metrics type", http.StatusBadRequest, 0)

		http.Error(w, "unknown metrics type", http.StatusBadRequest)

		return
	}

	resp, err := json.Marshal(m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	setContentType(w, "application/json")
	w.WriteHeader(http.StatusOK)

	size, err := w.Write(resp)
	if err != nil {
		h.logInfo(fmt.Sprintf("Error of write resp: %s", err.Error()), http.StatusInternalServerError, 0)
	}

	h.logInfo("Success get metrics", http.StatusOK, size)
}

func (h *Handler) AllValue(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), queryRepoTimeout)
	defer cancel()
	resp, err := h.repo.AllMetrics(ctx)
	if err != nil {
		h.logInfo(fmt.Sprintf("Failed to get all metrics: %s", err.Error()),
			http.StatusInternalServerError, 0)

		http.Error(w, fmt.Sprintf("failed to get all metrics: %s", err.Error()), http.StatusNotFound)

		return
	}

	setContentType(w, "text/html")
	w.WriteHeader(http.StatusOK)

	size, err := w.Write(resp)
	if err != nil {
		h.logInfo(fmt.Sprintf("Failed write metrics: %s", err.Error()),
			http.StatusInternalServerError, 0)
	}

	h.logInfo("Get all metrics", http.StatusOK, size)
}
