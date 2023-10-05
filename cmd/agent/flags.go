package main

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/vorotislav/alert-service/internal/settings/agent"
)

type Config struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

func parseFlags(sets *agent.Settings) {
	flag.StringVar(&sets.ServerAddress, "a", "localhost:8080", "server url")
	flag.IntVar(&sets.ReportInterval, "r", 10, "report interval, sec")
	flag.IntVar(&sets.PollInterval, "p", 2, "poll interval, sec")

	flag.Parse()

	cfg := Config{}
	if err := env.Parse(&cfg); err == nil {
		if cfg.Address != "" {
			sets.ServerAddress = cfg.Address
		}

		if cfg.PollInterval != 0 {
			sets.PollInterval = cfg.PollInterval
		}

		if cfg.ReportInterval != 0 {
			sets.ReportInterval = cfg.ReportInterval
		}
	}
}
