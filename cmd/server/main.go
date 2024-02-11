package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/vorotislav/alert-service/internal/grpc"

	"github.com/vorotislav/alert-service/internal/http"
	"github.com/vorotislav/alert-service/internal/repository"
	"github.com/vorotislav/alert-service/internal/settings/server"
	"github.com/vorotislav/alert-service/internal/signals"

	"go.uber.org/zap"
)

var (
	buildVersion = "N/A" //nolint:gochecknoglobals
	buildDate    = "N/A" //nolint:gochecknoglobals
	buildCommit  = "N/A" //nolint:gochecknoglobals
)

const serviceShutdownTimeout = 1 * time.Second

func main() {
	sets := server.Settings{}

	parseFlag(&sets)

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Printf("cannot create logger: %s", err.Error())

		return
	}

	defer func() {
		_ = logger.Sync()
	}()

	logger.Debug("Server starting...")
	logger.Debug(fmt.Sprintf("Build version: %s\nBuild date: %s\nBuild commit: %s\n",
		buildVersion, buildDate, buildCommit))
	logger.Debug("Current settings",
		zap.String("http address", sets.Address),
		zap.String("grpc address", sets.GAddress),
		zap.Bool("restore flag", *sets.Restore),
		zap.String("file path", sets.FileStoragePath),
		zap.String("database dsn", sets.DatabaseDSN),
		zap.String("hash key", sets.HashKey),
		zap.String("trusted subnet", sets.TrustedSubnet))

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

	gs := grpc.NewMetricServer(logger, repo, sets.GAddress)

	gserviceErrCh := make(chan error, 1)
	go func(errCh chan<- error) {
		defer close(errCh)

		if err := gs.Run(); err != nil {
			errCh <- err
		}
	}(gserviceErrCh)

	select {
	case err := <-serviceErrCh:
		if err != nil {
			logger.Error("service error", zap.Error(err))
			cancel()
		}
	case err := <-gserviceErrCh:
		if err != nil {
			logger.Error("grpc service error", zap.Error(err))
			cancel()
		}

	case <-ctx.Done():
		logger.Info("Server stopping...")

		ctxShutdown, ctxCancelShutdown := context.WithTimeout(context.Background(), serviceShutdownTimeout)

		if err := s.Stop(ctxShutdown); err != nil {
			logger.Error("cannot stop server", zap.Error(err))
		}

		if err := gs.Stop(ctxShutdown); err != nil {
			logger.Error("cannot stop grpc server", zap.Error(err))
		}

		defer ctxCancelShutdown()
	}
}
