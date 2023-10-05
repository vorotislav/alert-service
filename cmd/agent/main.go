package main

import (
	"context"
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

func main() {

	sets := agent.Settings{}

	parseFlags(&sets)

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Printf("cannot create logger: %s", err.Error())
		os.Exit(1)
	}

	defer logger.Sync()

	logger.Debug("Agent starting...")

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

	select {
	case <-ctx.Done():
		logger.Info("Agent stopping...")
		ctxShutdown, ctxCancelShutdown := context.WithTimeout(context.Background(), workerShutdownTimeout)
		worker.Stop(ctxShutdown)

		defer ctxCancelShutdown()
	}
}
