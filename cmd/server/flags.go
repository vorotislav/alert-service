package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/vorotislav/alert-service/internal/settings/server"

	"github.com/caarlos0/env/v6"
)

const (
	defaultAddress     = ":8080"
	defaultGAddress    = ":9090"
	defaultRestore     = true
	defaultInterval    = 300
	defaultStoragePath = "/tmp/metrics-db.json"
)

var errEmptyFilePath = errors.New("empty file path")

func parseFlag(sets *server.Settings) {
	var address string

	flag.StringVar(&address, "a", "", "address and port to run HTTP server")

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

	var trustedSubnet string

	flag.StringVar(&trustedSubnet, "t", "", "trusted subnet fot check incoming request")

	var gaddress string

	flag.StringVar(&gaddress, "g", "", "address and port to run gRPC server")

	flag.Parse()

	if err := env.Parse(sets); err != nil {
		return
	}

	if sets.Config == "" {
		sets.Config = configFile
	}

	fileCfg, _ := readConfigFile(sets.Config)

	if sets.Address == "" {
		sets.Address = getAddress(address, fileCfg.Address)
	}

	if sets.FileStoragePath == "" {
		sets.FileStoragePath = storagePath
	}

	if sets.StoreInterval == nil {
		storeInterval := getStoreInterval(interval, fileCfg.StoreInterval)
		sets.StoreInterval = &storeInterval
	}

	if sets.Restore == nil {
		sets.Restore = &restore
	}

	if sets.DatabaseDSN == "" {
		sets.DatabaseDSN = getDSN(databaseDSN, fileCfg.DatabaseDSN)
	}

	if sets.HashKey == "" {
		sets.HashKey = hashKey
	}

	if sets.CryptoKey == "" {
		sets.CryptoKey = getKey(cryptoKey, fileCfg.CryptoKey)
	}

	if sets.TrustedSubnet == "" {
		sets.TrustedSubnet = getSubnet(trustedSubnet, fileCfg.TrustedSubnet)
	}

	if sets.GAddress == "" {
		sets.GAddress = getGAddress(gaddress, fileCfg.GAddress)
	}
}

func readConfigFile(path string) (server.Settings, error) {
	if path == "" {
		return server.Settings{}, errEmptyFilePath
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return server.Settings{}, fmt.Errorf("cannot read config file: %w", err)
	}

	c := server.Settings{}

	err = json.Unmarshal(data, &c)
	if err != nil {
		return server.Settings{}, fmt.Errorf("cannot unmarshal config: %w", err)
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

func getGAddress(flag, conf string) string {
	if flag != "" {
		return flag
	}

	if conf != "" {
		return conf
	}

	return defaultGAddress
}

func getStoreInterval(flagInterval int, confInterval *int) int {
	if flagInterval > 0 {
		return flagInterval
	}

	if confInterval != nil {
		if *confInterval > 0 {
			return *confInterval
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

func getSubnet(flagSubnet, cfgSubnet string) string {
	if flagSubnet != "" {
		return flagSubnet
	}

	return cfgSubnet
}
