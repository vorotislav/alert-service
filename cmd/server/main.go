package main

import (
	"github.com/vorotislav/alert-service/internal/http"
	"log"
)

func main() {
	parseFlag()

	s := http.NewService(flagRunAddr)

	log.Fatal(s.Run())
}
