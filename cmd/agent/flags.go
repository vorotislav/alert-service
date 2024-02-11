package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/vorotislav/alert-service/internal/settings/agent"

	"github.com/caarlos0/env/v6"
)

const (
	defaultAddress        = "localhost:8080"
	defaultGAddress       = "localhost:9090"
	defaultPollInterval   = 2
	defaultReportInterval = 10
	defaultRateLimit      = 3
)

func parseFlags(sets *agent.Settings) { //nolint:gocognit,cyclop
	if err := env.Parse(sets); err == nil { //nolint:nestif
		if sets.ServerAddress == "" {
			flag.StringVar(&sets.ServerAddress, "a", "", "server url")
		}

		if sets.PollInterval == 0 {
			flag.IntVar(&sets.PollInterval, "p", 0, "poll interval, sec")
		}

		if sets.ReportInterval == 0 {
			flag.IntVar(&sets.ReportInterval, "r", 0, "report interval, sec")
		}

		if sets.HashKey == "" {
			flag.StringVar(&sets.HashKey, "k", "", "hash key")
		}

		if sets.RateLimit == 0 {
			flag.IntVar(&sets.RateLimit, "l", defaultRateLimit, "rate limit of worker pool")
		}

		if sets.CryptoKey == "" {
			flag.StringVar(&sets.CryptoKey, "crypto-key", "", "path to file with public key")
		}

		if sets.Config == "" {
			flag.StringVar(&sets.Config, "config", "", "path to config file")
		}

		if sets.GAddress == "" {
			flag.StringVar(&sets.GAddress, "g", "", "grpc server address")
		}
	}

	flag.Parse()

	var (
		cfg agent.Config
		err error
	)

	if sets.Config != "" {
		cfg, err = readConfigFile(sets.Config)
		if err != nil {
			return
		}
	}

	if sets.ServerAddress == "" {
		if cfg.Address != "" {
			sets.ServerAddress = cfg.Address
		} else {
			sets.ServerAddress = defaultAddress
		}
	}

	if sets.ReportInterval == 0 {
		if cfg.ReportInterval != "" {
			interval, err := strconv.Atoi(cfg.ReportInterval)
			if err != nil {
				sets.ReportInterval = defaultReportInterval
			} else {
				sets.ReportInterval = interval
			}
		} else {
			sets.ReportInterval = defaultReportInterval
		}
	}

	if sets.PollInterval == 0 {
		if cfg.PollInterval != "" {
			interval, err := strconv.Atoi(cfg.PollInterval)
			if err != nil {
				sets.PollInterval = defaultPollInterval
			} else {
				sets.PollInterval = interval
			}
		} else {
			sets.PollInterval = defaultPollInterval
		}
	}

	if sets.CryptoKey != "" {
		if cfg.CryptoKey != "" {
			sets.CryptoKey = cfg.CryptoKey
		}
	}

	if sets.GAddress == "" {
		if cfg.GAddress != "" {
			sets.GAddress = cfg.GAddress
		} else {
			sets.GAddress = defaultGAddress
		}
	}
}

func readConfigFile(path string) (agent.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return agent.Config{}, fmt.Errorf("cannot read config file: %w", err)
	}

	c := agent.Config{}

	err = json.Unmarshal(data, &c)
	if err != nil {
		return c, fmt.Errorf("cannot unmarshal config: %w", err)
	}

	return c, nil
}
