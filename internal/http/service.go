package http

import (
	"context"
	"github.com/vorotislav/alert-service/internal/http/handlers"
	"github.com/vorotislav/alert-service/internal/repository"
	"github.com/vorotislav/alert-service/internal/settings/server"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vorotislav/alert-service/internal/http/middlewares"
	"go.uber.org/zap"
)

type Service struct {
	logger  *zap.Logger
	server  *http.Server
	handler *handlers.Handler
	repo    repository.Repository
}

func NewService(
	_ context.Context,
	log *zap.Logger,
	set *server.Settings,
	repo repository.Repository,
) (*Service, error) {
	r := chi.NewRouter()

	r.Use(middlewares.New(log))
	r.Use(middlewares.CompressMiddleware)
	if set.HashKey != "" {
		r.Use(middlewares.Hash(log, set.HashKey))
	}

	handler := handlers.NewHandler(log, repo)

	r.Route("/updates", func(r chi.Router) {
		r.Post("/", handler.Updates)
	})
	r.Route("/update", func(r chi.Router) {
		r.Route("/{metricType}", func(r chi.Router) {
			r.Route("/{metricName}", func(r chi.Router) {
				r.Post("/{metricValue}", handler.Update)
			})
		})

		r.Post("/", handler.UpdateJSON)
	})

	r.Route("/value", func(r chi.Router) {
		r.Route("/{metricType}", func(r chi.Router) {
			r.Get("/{metricName}", handler.Value)
		})

		r.Post("/", handler.ValueJSON)
	})

	r.Route("/ping", func(r chi.Router) {
		r.Get("/", handler.Ping)
	})

	r.Get("/", handler.AllValue)

	hs := &http.Server{
		Addr:    set.Address,
		Handler: r,
	}

	return &Service{
		logger:  log.With(zap.String("package", "service")),
		server:  hs,
		repo:    repo,
		handler: handler,
	}, nil
}

func (s *Service) Run() error {
	s.logger.Info("Running server on", zap.String("address", s.server.Addr))
	return s.server.ListenAndServe()
}

func (s *Service) Stop(ctx context.Context) error {
	s.logger.Debug("Stopping service")
	if err := s.repo.Stop(ctx); err != nil {
		s.logger.Error("error of repo stop", zap.Error(err))
	}

	return s.server.Shutdown(ctx)
}
