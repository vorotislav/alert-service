package main

import (
	"context"
	"github.com/vorotislav/alert-service/internal/http"
	"github.com/vorotislav/alert-service/internal/settings"
	"go.uber.org/zap"
	"log"
)

func main() {
	sets := settings.Settings{}

	parseFlag(&sets)

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	defer logger.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s, err := http.NewService(ctx, logger, &sets)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(s.Run())
}
