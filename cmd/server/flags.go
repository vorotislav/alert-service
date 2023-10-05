package main

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/vorotislav/alert-service/internal/settings/server"
)

type Config struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   *int   `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         *bool  `env:"RESTORE"`
}

func parseFlag(sets *server.Settings) {
	flag.StringVar(&sets.Address, "a", ":8080", "address and port to run server")
	flag.IntVar(&sets.StoreInterval, "i", 300, "store interval, sec")
	flag.StringVar(&sets.FileStoragePath, "f", "/tmp/metrics-db.json", "file storage path")
	flag.BoolVar(&sets.Restore, "r", true, "restore old metrics")

	flag.Parse()

	cfg := Config{}
	if err := env.Parse(&cfg); err == nil {
		if cfg.Address != "" {
			sets.Address = cfg.Address
		}
		if cfg.FileStoragePath != "" {
			sets.FileStoragePath = cfg.FileStoragePath
		}
		if cfg.Restore != nil {
			sets.Restore = *cfg.Restore
		}
		if cfg.StoreInterval != nil {
			sets.StoreInterval = *cfg.StoreInterval
		}
	}
}
