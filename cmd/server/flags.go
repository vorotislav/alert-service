package main

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/vorotislav/alert-service/internal/settings/server"
)

func parseFlag(sets *server.Settings) {
	var address string
	flag.StringVar(&address, "a", ":8080", "address and port to run server")
	var restore bool
	flag.BoolVar(&restore, "r", true, "restore old metrics")
	var interval int
	flag.IntVar(&interval, "i", 300, "store interval, sec")

	var databaseDSN string
	flag.StringVar(&databaseDSN, "d", "", "database dsn")

	var storagePath string
	flag.StringVar(&storagePath, "f", "/tmp/metrics-db.json", "file localstorage path")

	flag.Parse()

	if err := env.Parse(sets); err == nil {
		if sets.Address == "" {
			sets.Address = address
		}
		if sets.FileStoragePath == "" {
			sets.FileStoragePath = storagePath
		}
		if sets.StoreInterval == nil {
			sets.StoreInterval = &interval
		}
		if sets.Restore == nil {
			sets.Restore = &restore
		}
		if sets.DatabaseDSN == "" {
			sets.DatabaseDSN = databaseDSN
		}
	}

	// user=postgres password=postgres host=127.0.0.1 port=5432 dbname=alert_service pool_max_conns=10
	//flag.Parse()
}
