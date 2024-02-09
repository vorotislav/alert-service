// Package http представляет сервис для создания обработчика, настройки маршрутов и запуска и остановки http сервера.
package http

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/vorotislav/alert-service/internal/http/handlers"
	"github.com/vorotislav/alert-service/internal/http/middlewares"
	"github.com/vorotislav/alert-service/internal/repository"
	"github.com/vorotislav/alert-service/internal/settings/server"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

const (
	defaultReadHeaderTimeout = time.Second
)

// Service сущность сервиса. Хранит логгер, http-сервер, обработчик и репозиторий.
type Service struct {
	logger  *zap.Logger
	server  *http.Server
	handler *handlers.Handler
	repo    repository.Repository
}

// NewService конструктор для Service.
func NewService(
	_ context.Context,
	log *zap.Logger,
	set *server.Settings,
	repo repository.Repository,
) (*Service, error) {
	r := chi.NewRouter()

	r.Use(middlewares.New(log))

	// проверяем, что есть доверенная сеть для агентов
	if set.TrustedSubnet != "" {
		r.Use(middlewares.CheckSenderIP(log, set.TrustedSubnet))
	}

	if set.HashKey != "" {
		r.Use(middlewares.Hash(log, set.HashKey))
	}

	if set.CryptoKey != "" {
		r.Use(middlewares.DecryptMiddleware(log, set.CryptoKey))
	}

	r.Use(middlewares.CompressMiddleware)

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
	r.HandleFunc("/debug/pprof/heap", pprof.Index)

	hs := &http.Server{
		Addr:              set.Address,
		Handler:           r,
		ReadHeaderTimeout: defaultReadHeaderTimeout,
	}

	return &Service{
		logger:  log.With(zap.String("package", "service")),
		server:  hs,
		repo:    repo,
		handler: handler,
	}, nil
}

// Run запускает http-сервер.
func (s *Service) Run() error {
	s.logger.Info("Running server on", zap.String("address", s.server.Addr))

	return s.server.ListenAndServe() //nolint:wrapcheck
}

// Stop останавливает http-сервер.
func (s *Service) Stop(ctx context.Context) error {
	s.logger.Debug("Stopping service")

	if err := s.repo.Stop(ctx); err != nil {
		s.logger.Error("error of repo stop", zap.Error(err))
	}

	err := s.server.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	return nil
}
