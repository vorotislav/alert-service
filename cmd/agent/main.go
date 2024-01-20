// Модуль main запускает основное приложение по сбору метрик и отправки на сервер.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/vorotislav/alert-service/internal/http/client"
	"github.com/vorotislav/alert-service/internal/metrics"
	"github.com/vorotislav/alert-service/internal/settings/agent"
	"github.com/vorotislav/alert-service/internal/signals"

	"go.uber.org/zap"
)

const workerShutdownTimeout = 1 * time.Second

var (
	BuildVersion = "N/A"
	BuildDate    = "N/A"
	BuildCommit  = "N/A"
)

func main() {

	sets := agent.Settings{}

	parseFlags(&sets)

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Printf("cannot create logger: %s", err.Error())
		return
	}

	defer logger.Sync()

	logger.Debug("Agent starting...")
	logger.Info(fmt.Sprintf("Build version: %s\nBuild date: %s\nBuild commit: %s\n",
		BuildVersion, BuildDate, BuildCommit))
	logger.Debug("Current settings",
		zap.String("server address", sets.ServerAddress),
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

	wc := client.NewClient(logger, &sets)

	worker := metrics.NewWorker(logger, &sets, wc)
	worker.Start(ctx)

	<-ctx.Done()
	logger.Info("Agent stopping...")
	ctxShutdown, ctxCancelShutdown := context.WithTimeout(context.Background(), workerShutdownTimeout)
	worker.Stop(ctxShutdown)

	defer ctxCancelShutdown()
}
