package main

import (
	"context"
	"github.com/vorotislav/alert-service/internal/http"
	"github.com/vorotislav/alert-service/internal/settings/server"
	"github.com/vorotislav/alert-service/internal/signals"
	"go.uber.org/zap"
	"log"
	"os"
	"time"
)

const serviceShutdownTimeout = 1 * time.Second

func main() {
	sets := server.Settings{}

	parseFlag(&sets)

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Printf("cannot create logger: %s", err.Error())

		os.Exit(1)
	}

	defer logger.Sync()

	logger.Debug("Server starting...")

	ctx, cancel := context.WithCancel(context.Background())
	oss := signals.NewOSSignals(ctx)

	oss.Subscribe(func(sig os.Signal) {
		logger.Info("Stopping by OS Signal...",
			zap.String("signal", sig.String()))

		cancel()
	})

	s, err := http.NewService(ctx, logger, &sets)
	if err != nil {
		log.Fatal(err)
	}

	serviceErrCh := make(chan error, 1)
	go func(errCh chan<- error) {
		defer close(errCh)

		if err := s.Run(); err != nil {
			errCh <- err
		}
	}(serviceErrCh)

	select {
	case err := <-serviceErrCh:
		if err != nil {
			logger.Error("service error", zap.Error(err))
			cancel()
		}
	case <-ctx.Done():
		logger.Info("Server stopping...")
		ctxShutdown, ctxCancelShutdown := context.WithTimeout(context.Background(), serviceShutdownTimeout)

		_ = s.Stop(ctxShutdown)

		defer ctxCancelShutdown()
	}
}
