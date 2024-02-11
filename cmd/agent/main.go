// Модуль main запускает основное приложение по сбору метрик и отправки на сервер.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	grpcClient "github.com/vorotislav/alert-service/internal/grpc/client"
	httpClient "github.com/vorotislav/alert-service/internal/http/client"
	"github.com/vorotislav/alert-service/internal/metrics"
	"github.com/vorotislav/alert-service/internal/settings/agent"
	"github.com/vorotislav/alert-service/internal/signals"

	"go.uber.org/zap"
)

const workerShutdownTimeout = 1 * time.Second

var (
	buildVersion = "N/A" //nolint:gochecknoglobals
	buildDate    = "N/A" //nolint:gochecknoglobals
	buildCommit  = "N/A" //nolint:gochecknoglobals
)

func main() {
	sets := agent.Settings{}

	parseFlags(&sets)

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Printf("cannot create logger: %s", err.Error())

		return
	}

	defer func() {
		_ = logger.Sync()
	}()

	logger.Debug("Agent starting...")
	logger.Info(fmt.Sprintf("Build version: %s\nBuild date: %s\nBuild commit: %s\n",
		buildVersion, buildDate, buildCommit))
	logger.Debug("Current settings",
		zap.String("server address", sets.ServerAddress),
		zap.String("grpc server address", sets.GAddress),
		zap.Int("report interval", sets.ReportInterval),
		zap.Int("poll interval", sets.PollInterval),
		zap.Int("rate limit", sets.RateLimit),
		zap.String("hash key", sets.HashKey))

	ctx, cancel := context.WithCancel(context.Background())
	oss := signals.NewOSSignals(ctx)

	oss.Subscribe(func(sig os.Signal) {
		logger.Info("Stopping by OS Signal...",
			zap.String("signal", sig.String()))

		cancel()
	})

	var (
		wc metrics.Client
	)

	if sets.GAddress != "" {
		wc, err = grpcClient.NewClient(logger, sets.GAddress)
	} else {
		wc, err = httpClient.NewClient(logger, &sets)
	}

	if err != nil {
		logger.Error("cannot create client for sending metrics", zap.Error(err))

		return
	}

	worker := metrics.NewWorker(logger, &sets, wc)
	worker.Start(ctx)

	<-ctx.Done()
	logger.Info("Agent stopping...")

	ctxShutdown, ctxCancelShutdown := context.WithTimeout(context.Background(), workerShutdownTimeout)

	worker.Stop(ctxShutdown)

	defer ctxCancelShutdown()
}
