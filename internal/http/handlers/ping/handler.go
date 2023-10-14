package ping

import (
	"context"
	"go.uber.org/zap"
	"net/http"
)

type Repository interface {
	Ping(ctx context.Context) error
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

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	err := h.repo.Ping(r.Context())
	if err != nil {
		http.Error(w, "repository is not available", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
}
