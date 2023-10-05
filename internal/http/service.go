package http

import (
	"context"
	"fmt"
	"github.com/vorotislav/alert-service/internal/http/handlers/update"
	"github.com/vorotislav/alert-service/internal/http/handlers/value"
	"github.com/vorotislav/alert-service/internal/settings/server"
	"net/http"

	"github.com/vorotislav/alert-service/internal/http/middlewares"
	"github.com/vorotislav/alert-service/internal/storage"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Service struct {
	logger        *zap.Logger
	server        *http.Server
	updateHandler *update.Handler
	valueHandler  *value.Handler
	store         *storage.MemStorage
}

func NewService(ctx context.Context, log *zap.Logger, set *server.Settings) (*Service, error) {
	r := chi.NewRouter()

	r.Use(middlewares.New(log))
	r.Use(middlewares.CompressMiddleware)

	store, err := storage.NewMemStorage(ctx, log, set)
	if err != nil {
		return nil, fmt.Errorf("cannot create service: %w", err)
	}

	updateMetricHandler := update.NewHandler(log, store)

	valueMetricHandler := value.NewHandler(log, store)

	r.Route("/update", func(r chi.Router) {
		r.Route("/{metricType}", func(r chi.Router) {
			r.Route("/{metricName}", func(r chi.Router) {
				r.Post("/{metricValue}", updateMetricHandler.Update)
			})
		})

		r.Post("/", updateMetricHandler.UpdateJSON)
	})

	r.Route("/value", func(r chi.Router) {
		r.Route("/{metricType}", func(r chi.Router) {
			r.Get("/{metricName}", valueMetricHandler.Value)
		})

		r.Post("/", valueMetricHandler.ValueJSON)
	})

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		resp, err := valueMetricHandler.AllMetrics()
		if err != nil {
			log.Info("Failed to get all counter metrics",
				zap.Int("status code", http.StatusInternalServerError),
				zap.Int("size", 0))

			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "text/html")
			return
		}

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)

		size, err := w.Write(resp)
		if err != nil {
			log.Info("Failed to get all metrics",
				zap.Int("status code", http.StatusInternalServerError),
				zap.Int("size", 0),
				zap.String("error", err.Error()))
		}

		log.Info("Get all metrics",
			zap.Int("status code", http.StatusOK),
			zap.Int("size", size))
	})

	hs := &http.Server{
		Addr:    set.Address,
		Handler: r,
	}

	return &Service{
		logger:        log.With(zap.String("package", "service")),
		server:        hs,
		store:         store,
		updateHandler: updateMetricHandler,
		valueHandler:  valueMetricHandler,
	}, nil
}

func (s *Service) Run() error {
	s.logger.Info("Running server on", zap.String("address", s.server.Addr))
	return s.server.ListenAndServe()
}

func (s *Service) Stop(ctx context.Context) error {
	s.logger.Debug("Stopping service")
	if err := s.store.Stop(ctx); err != nil {
		s.logger.Error("error of store stop", zap.Error(err))
	}
	return s.server.Shutdown(ctx)
}
