package main

import (
	"context"
	"github.com/vorotislav/alert-service/internal/http"
	"github.com/vorotislav/alert-service/internal/repository"
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
	logger.Debug("Restore flag: ", zap.Bool("flag", *sets.Restore))
	logger.Debug("File path: ", zap.String("path", sets.FileStoragePath))

	ctx, cancel := context.WithCancel(context.Background())
	oss := signals.NewOSSignals(ctx)

	oss.Subscribe(func(sig os.Signal) {
		logger.Info("Stopping by OS Signal...",
			zap.String("signal", sig.String()))

		cancel()
	})

	repo, err := repository.NewRepository(ctx, logger, &sets)

	if err != nil {
		logger.Error("cannot create repository", zap.Error(err))

		return
	}

	s, err := http.NewService(ctx, logger, &sets, repo)
	if err != nil {
		logger.Error("cannot create http service", zap.Error(err))

		return
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

		if err := s.Stop(ctxShutdown); err != nil {
			logger.Error("cannot stop server", zap.Error(err))
		}

		defer ctxCancelShutdown()
	}
}
