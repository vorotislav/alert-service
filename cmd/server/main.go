package main

import (
	"github.com/vorotislav/alert-service/internal/http"
	"log"
)

func main() {
	s := http.NewService()

	log.Fatal(s.Run())
}
