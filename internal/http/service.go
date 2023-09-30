package http

import (
	"fmt"
	"github.com/vorotislav/alert-service/internal/http/handlers/update"
	"github.com/vorotislav/alert-service/internal/http/handlers/value"
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
}

func NewService(log *zap.Logger, addr string) *Service {
	r := chi.NewRouter()

	r.Use(middlewares.New(log))

	store := storage.NewMemStorage()

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
			log.Info("Failed to get all counter metric",
				zap.Int("status code", http.StatusInternalServerError),
				zap.Int("size", 0))

			http.Error(w, "", http.StatusInternalServerError)
		}

		size, err := w.Write(resp)
		if err != nil {
			log.Info("Failed to get all metric",
				zap.Int("status code", http.StatusInternalServerError),
				zap.Int("size", 0),
				zap.String("error", err.Error()))
		}
		w.WriteHeader(http.StatusOK)

		log.Info("Get all metric",
			zap.Int("status code", http.StatusOK),
			zap.Int("size", size))

		w.Header().Set("Content-Type", "text/plain")
		w.Header().Add("Content-Type", "charset=utf-8")
	})

	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	return &Service{
		logger:        log,
		server:        server,
		updateHandler: updateMetricHandler,
		valueHandler:  valueMetricHandler,
	}
}

func (s *Service) Run() error {
	fmt.Println("Running server on", s.server.Addr)
	return s.server.ListenAndServe()
}
