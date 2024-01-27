package main

import (
	"flag"

	"github.com/vorotislav/alert-service/internal/settings/server"

	"github.com/caarlos0/env/v6"
)

const (
	defaultAddress     = ":8080"
	defaultRestore     = true
	defaultInterval    = 300
	defaultStoragePath = "/tmp/metrics-db.json"
)

func parseFlag(sets *server.Settings) {
	var address string

	flag.StringVar(&address, "a", defaultAddress, "address and port to run server")

	var restore bool

	flag.BoolVar(&restore, "r", defaultRestore, "restore old metrics")

	var interval int

	flag.IntVar(&interval, "i", defaultInterval, "store interval, sec")

	var databaseDSN string

	flag.StringVar(&databaseDSN, "d", "", "database dsn")

	var storagePath string

	flag.StringVar(&storagePath, "f", defaultStoragePath, "file localstorage path")

	var hashKey string

	flag.StringVar(&hashKey, "k", "", "hash key")

	flag.Parse()

	if err := env.Parse(sets); err != nil {
		return
	}

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

	if sets.HashKey == "" {
		sets.HashKey = hashKey
	}
}
