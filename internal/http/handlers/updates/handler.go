package updates

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/vorotislav/alert-service/internal/model"
	"go.uber.org/zap"
	"net/http"
)

type Repository interface {
	UpdateMetrics(ctx context.Context, metrics []model.Metrics) error
}

type Handler struct {
	log  *zap.Logger
	repo Repository
}

func NewHandler(log *zap.Logger, repo Repository) *Handler {
	return &Handler{
		log:  log,
		repo: repo,
	}
}

func (h *Handler) Updates(w http.ResponseWriter, r *http.Request) {
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		h.log.Info("Failed to update metrics",
			zap.String("unknown Content-Type", contentType),
			zap.Int("status code", http.StatusBadRequest),
			zap.Int("size", 0))

		http.Error(w, "unknown Content-Type", http.StatusBadRequest)

		return
	}

	metrics := make([]model.Metrics, 0)
	err := json.NewDecoder(r.Body).Decode(&metrics)
	if err != nil {
		h.log.Info("Failed to update metrics",
			zap.String("cannot decode metrics", err.Error()),
			zap.Int("status code", http.StatusBadRequest),
			zap.Int("size", 0))

		http.Error(w, "cannot decode body", http.StatusBadRequest)

		return
	}

	if err := h.repo.UpdateMetrics(r.Context(), metrics); err != nil {
		h.log.Info("Failed to update metrics",
			zap.String("cannot update metrics", err.Error()),
			zap.Int("status code", http.StatusBadRequest),
			zap.Int("size", 0))

		http.Error(w, fmt.Sprintf("cannot update metrics: %s", err.Error()), http.StatusBadRequest)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	size, err := w.Write([]byte(`{}`))
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
