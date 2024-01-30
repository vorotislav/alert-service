package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/vorotislav/alert-service/internal/settings/server"

	"github.com/caarlos0/env/v6"
)

const (
	defaultAddress     = ":8080"
	defaultRestore     = true
	defaultInterval    = 300
	defaultStoragePath = "/tmp/metrics-db.json"
)

var errEmptyFilePath = errors.New("empty file path")

func parseFlag(sets *server.Settings) {
	var address string

	flag.StringVar(&address, "a", "", "address and port to run server")

	var restore bool

	flag.BoolVar(&restore, "r", defaultRestore, "restore old metrics")

	var interval int

	flag.IntVar(&interval, "i", 0, "store interval, sec")

	var databaseDSN string

	flag.StringVar(&databaseDSN, "d", "", "database dsn")

	var storagePath string

	flag.StringVar(&storagePath, "f", defaultStoragePath, "file localstorage path")

	var hashKey string

	flag.StringVar(&hashKey, "k", "", "hash key")

	var cryptoKey string

	flag.StringVar(&cryptoKey, "crypto-key", "", "path to file with private key")

	var configFile string

	flag.StringVar(&configFile, "config", "", "path to config file")

	flag.Parse()

	if err := env.Parse(sets); err != nil {
		return
	}

	if sets.Config == "" {
		sets.Config = configFile
	}

	cfg, _ := readConfigFile(sets.Config)

	if sets.Address == "" {
		sets.Address = getAddress(address, cfg.Address)
	}

	if sets.FileStoragePath == "" {
		sets.FileStoragePath = storagePath
	}

	if sets.StoreInterval == nil {
		storeInterval := getStoreInterval(interval, cfg.StoreInterval)
		sets.StoreInterval = &storeInterval
	}

	if sets.Restore == nil {
		sets.Restore = &restore
	}

	if sets.DatabaseDSN == "" {
		sets.DatabaseDSN = getDSN(databaseDSN, cfg.DatabaseDsn)
	}

	if sets.HashKey == "" {
		sets.HashKey = hashKey
	}

	if sets.CryptoKey == "" {
		sets.CryptoKey = getKey(cryptoKey, cfg.CryptoKey)
	}
}

func readConfigFile(path string) (server.Config, error) {
	if path == "" {
		return server.Config{}, errEmptyFilePath
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return server.Config{}, fmt.Errorf("cannot read config file: %w", err)
	}

	c := server.Config{}
	err = json.Unmarshal(data, &c)
	if err != nil {
		return server.Config{}, fmt.Errorf("cannot unmarshal config: %w", err)
	}

	return c, nil
}

func getAddress(flag, conf string) string {
	if flag != "" {
		return flag
	}

	if conf != "" {
		return conf
	}

	return defaultAddress
}

func getStoreInterval(flagInterval int, confInterval *string) int {
	if flagInterval > 0 {
		return flagInterval
	}

	if confInterval != nil && *confInterval != "" {
		interval, err := strconv.Atoi(*confInterval)
		if err != nil {
			return defaultInterval
		}

		if interval > 0 {
			return interval
		}
	}

	return defaultInterval
}

func getDSN(flagDSN, cfgDSN string) string {
	if flagDSN != "" {
		return flagDSN
	}

	return cfgDSN
}

func getKey(flagKey, cfgKey string) string {
	if flagKey != "" {
		return flagKey
	}

	return cfgKey
}
