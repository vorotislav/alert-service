package main

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/vorotislav/alert-service/internal/settings"
)

var (
	flagRunAddr         string
	flagStoreInterval   int
	flagFileStoragePath string
	flagRestore         bool
)

type Config struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   *int   `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         *bool  `env:"RESTORE"`
}

func parseFlag(sets *settings.Settings) {
	flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
	flag.IntVar(&flagStoreInterval, "i", 300, "store interval, sec")
	flag.StringVar(&flagFileStoragePath, "f", "/tmp/metrics-db.json", "file storage path")
	flag.BoolVar(&flagRestore, "r", true, "restore old metrics")

	flag.Parse()

	cfg := Config{}
	if err := env.Parse(&cfg); err == nil {
		if cfg.Address != "" {
			flagRunAddr = cfg.Address
		}
		if cfg.FileStoragePath != "" {
			flagFileStoragePath = cfg.FileStoragePath
		}
		if cfg.Restore != nil {
			flagRestore = *cfg.Restore
		}
		if cfg.StoreInterval != nil {
			flagStoreInterval = *cfg.StoreInterval
		}
	}

	sets.Address = flagRunAddr
	sets.Restore = flagRestore
	sets.StoreInterval = flagStoreInterval
	sets.FileStoragePath = flagFileStoragePath
}
