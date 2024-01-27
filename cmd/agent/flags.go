package main

import (
	"flag"

	"github.com/vorotislav/alert-service/internal/settings/agent"

	"github.com/caarlos0/env/v6"
)

const (
	defaultPollInterval   = 2
	defaultReportInterval = 10
	defaultRateLimit      = 3
)

func parseFlags(sets *agent.Settings) {
	if err := env.Parse(sets); err == nil { //nolint:nestif
		if sets.ServerAddress == "" {
			flag.StringVar(&sets.ServerAddress, "a", "localhost:8080", "server url")
		}

		if sets.PollInterval == 0 {
			flag.IntVar(&sets.PollInterval, "p", defaultPollInterval, "poll interval, sec")
		}

		if sets.ReportInterval == 0 {
			flag.IntVar(&sets.ReportInterval, "r", defaultReportInterval, "report interval, sec")
		}

		if sets.HashKey == "" {
			flag.StringVar(&sets.HashKey, "k", "", "hash key")
		}

		if sets.RateLimit == 0 {
			flag.IntVar(&sets.RateLimit, "l", defaultRateLimit, "rate limit of worker pool")
		}
	}

	flag.Parse()
}
