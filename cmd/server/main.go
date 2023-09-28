package main

import (
	"github.com/vorotislav/alert-service/internal/http"
	"go.uber.org/zap"
	"log"
)

func main() {
	parseFlag()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	defer logger.Sync()

	s := http.NewService(logger, flagRunAddr)

	log.Fatal(s.Run())
}
