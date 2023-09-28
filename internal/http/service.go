package http

import (
	"fmt"
	"net/http"

	"github.com/vorotislav/alert-service/internal/http/handlers/metrics/counter"
	"github.com/vorotislav/alert-service/internal/http/handlers/metrics/gauge"
	"github.com/vorotislav/alert-service/internal/http/middlewares"
	"github.com/vorotislav/alert-service/internal/storage"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Service struct {
	logger         *zap.Logger
	server         *http.Server
	counterHandler *counter.Handler
	gaugeHandler   *gauge.Handler
}

func NewService(log *zap.Logger, addr string) *Service {
	r := chi.NewRouter()

	r.Use(middlewares.New(log))

	store := storage.NewMemStorage()

	counterMetricsHandler := counter.NewHandler(log, store)

	gaugeMetricsHandler := gauge.NewHandler(log, store)

	r.Route("/update", func(r chi.Router) {
		r.Route("/counter", func(r chi.Router) {
			r.Route("/{metricName}", func(r chi.Router) {
				r.Post("/{metricValue}", counterMetricsHandler.Update)
			})
		})

		r.Route("/gauge", func(r chi.Router) {
			r.Route("/{metricName}", func(r chi.Router) {
				r.Post("/{metricValue}", gaugeMetricsHandler.Update)
			})
		})

		r.Route("/{metricType}", func(r chi.Router) {
			r.Route("/{metricName}", func(r chi.Router) {
				r.Post("/{metricValue}", func(writer http.ResponseWriter, request *http.Request) {
					http.Error(writer, "", http.StatusBadRequest)
				})
			})
		})
	})

	r.Route("/value", func(r chi.Router) {
		r.Route("/counter", func(r chi.Router) {
			r.Route("/{metricName}", func(r chi.Router) {
				r.Get("/", counterMetricsHandler.Value)
			})
		})

		r.Route("/gauge", func(r chi.Router) {
			r.Route("/{metricName}", func(r chi.Router) {
				r.Get("/", gaugeMetricsHandler.Value)
			})

		})
		r.Route("/{metricType}", func(r chi.Router) {
			r.Get("/{metricName}", func(writer http.ResponseWriter, request *http.Request) {
				http.Error(writer, "", http.StatusBadRequest)
			})
		})

	})

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		respCounter, err := counterMetricsHandler.AllMetrics()
		if err != nil {
			log.Info("Failed to get all counter metric",
				zap.Int("status code", http.StatusInternalServerError),
				zap.Int("size", 0))

			http.Error(w, "", http.StatusInternalServerError)
		}

		respGauge, err := gaugeMetricsHandler.AllMetrics()
		if err != nil {
			log.Info("Failed to get all gauge metric",
				zap.Int("status code", http.StatusInternalServerError),
				zap.Int("size", 0))

			http.Error(w, "", http.StatusInternalServerError)
		}

		resp := append(respCounter, respGauge...)

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
		logger:         log,
		server:         server,
		counterHandler: counterMetricsHandler,
		gaugeHandler:   gaugeMetricsHandler,
	}
}

func (s *Service) Run() error {
	fmt.Println("Running server on", s.server.Addr)
	return s.server.ListenAndServe()
}
