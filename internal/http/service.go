package http

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"

	"github.com/vorotislav/alert-service/internal/http/handlers/metrics/counter"
	"github.com/vorotislav/alert-service/internal/http/handlers/metrics/gauge"
	"github.com/vorotislav/alert-service/internal/storage"
)

type Service struct {
	server         *http.Server
	counterHandler *counter.Handler
	gaugeHandler   *gauge.Handler
}

func NewService(addr string) *Service {
	r := chi.NewRouter()

	store := storage.NewMemStorage()

	counterMetricsHandler := counter.NewHandler(store)

	gaugeMetricsHandler := gauge.NewHandler(store)

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
			http.Error(w, "", http.StatusInternalServerError)
		}

		respGauge, err := gaugeMetricsHandler.AllMetrics()
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
		}

		resp := append(respCounter, respGauge...)

		w.Write(resp)
		w.WriteHeader(http.StatusOK)

		w.Header().Set("Content-Type", "text/plain")
		w.Header().Add("Content-Type", "charset=utf-8")
	})

	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	return &Service{
		server:         server,
		counterHandler: counterMetricsHandler,
		gaugeHandler:   gaugeMetricsHandler,
	}
}

func (s *Service) Run() error {
	fmt.Println("Running server on", s.server.Addr)
	return s.server.ListenAndServe()
}
