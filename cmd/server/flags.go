package main

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/vorotislav/alert-service/internal/settings/server"
)

func parseFlag(sets *server.Settings) {
	if err := env.Parse(sets); err == nil {
		if sets.Address == "" {
			flag.StringVar(&sets.Address, "a", ":8080", "address and port to run server")
		}
		if sets.FileStoragePath == "" {
			flag.StringVar(&sets.FileStoragePath, "f", "/tmp/metrics-db.json", "file localstorage path")
		}
		if sets.StoreInterval == nil {
			var interval int
			flag.IntVar(&interval, "i", 300, "store interval, sec")
			sets.StoreInterval = &interval
		}
		if sets.Restore == nil {
			var restore bool
			flag.BoolVar(&restore, "r", true, "restore old metrics")
			sets.Restore = &restore
		}
		if sets.DatabaseDSN == "" {
			flag.StringVar(&sets.DatabaseDSN, "d", "", "database dsn")
		}
	}

	// user=postgres password=postgres host=127.0.0.1 port=5432 dbname=alert_service pool_max_conns=10
	flag.Parse()
}
