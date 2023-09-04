package http

import (
	"net/http"

	"github.com/vorotislav/alert-service/internal/http/handlers/metrics/counter"
	"github.com/vorotislav/alert-service/internal/http/handlers/metrics/gauge"
	"github.com/vorotislav/alert-service/internal/storage"
)

type Service struct {
	server http.Server
}

func NewService() *Service {
	mux := http.NewServeMux()

	store := storage.NewMemStorage()

	counterMetricsHandler := counter.NewHandler(store)

	gaugeMetricsHandler := gauge.NewHandler(store)

	mux.Handle("/update/counter/", counterMetricsHandler)
	mux.Handle("/update/gauge/", gaugeMetricsHandler)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	return &Service{
		server: server,
	}
}

func (s *Service) Run() error {
	return s.server.ListenAndServe()
}
