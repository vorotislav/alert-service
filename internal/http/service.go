package http

import (
	"context"
	"github.com/vorotislav/alert-service/internal/http/handlers/ping"
	"github.com/vorotislav/alert-service/internal/http/handlers/update"
	"github.com/vorotislav/alert-service/internal/http/handlers/updates"
	"github.com/vorotislav/alert-service/internal/http/handlers/value"
	"github.com/vorotislav/alert-service/internal/model"
	"github.com/vorotislav/alert-service/internal/settings/server"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vorotislav/alert-service/internal/http/middlewares"
	"go.uber.org/zap"
)

type Repository interface {
	UpdateCounter(ctx context.Context, name string, value int64) (int64, error)
	GetCounterValue(ctx context.Context, name string) (int64, error)
	AllCounterMetrics(ctx context.Context) ([]byte, error)
	UpdateGauge(ctx context.Context, name string, value float64) (float64, error)
	GetGaugeValue(ctx context.Context, name string) (float64, error)
	AllGaugeMetrics(ctx context.Context) ([]byte, error)
	Ping(ctx context.Context) error
	Stop(ctx context.Context) error
	UpdateMetrics(ctx context.Context, metrics []model.Metrics) error
}

type Service struct {
	logger         *zap.Logger
	server         *http.Server
	updateHandler  *update.Handler
	valueHandler   *value.Handler
	pingHandler    *ping.Handler
	updatesHandler *updates.Handler
	repo           Repository
}

func NewService(_ context.Context, log *zap.Logger, set *server.Settings, repo Repository) (*Service, error) {
	r := chi.NewRouter()

	r.Use(middlewares.New(log))
	r.Use(middlewares.CompressMiddleware)

	updateMetricHandler := update.NewHandler(log, repo)

	valueMetricHandler := value.NewHandler(log, repo)

	pingHandler := ping.NewHandler(log, repo)

	updatesMetricHandler := updates.NewHandler(log, repo)

	r.Route("/updates", func(r chi.Router) {
		r.Post("/", updatesMetricHandler.Updates)
	})
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

	r.Route("/ping", func(r chi.Router) {
		r.Get("/", pingHandler.Ping)
	})

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		resp, err := valueMetricHandler.AllMetrics(r.Context())
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
		logger:         log.With(zap.String("package", "service")),
		server:         hs,
		repo:           repo,
		updateHandler:  updateMetricHandler,
		valueHandler:   valueMetricHandler,
		pingHandler:    pingHandler,
		updatesHandler: updatesMetricHandler,
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
