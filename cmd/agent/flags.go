package main

import (
	"flag"

	"github.com/caarlos0/env/v6"
	"github.com/vorotislav/alert-service/internal/settings/agent"
)

func parseFlags(sets *agent.Settings) {
	if err := env.Parse(sets); err == nil {
		if sets.ServerAddress == "" {
			flag.StringVar(&sets.ServerAddress, "a", "localhost:8080", "server url")
		}

		if sets.PollInterval == 0 {
			flag.IntVar(&sets.PollInterval, "p", 2, "poll interval, sec")
		}

		if sets.ReportInterval == 0 {
			flag.IntVar(&sets.ReportInterval, "r", 10, "report interval, sec")
		}

		if sets.HashKey == "" {
			flag.StringVar(&sets.HashKey, "k", "", "hash key")
		}

		if sets.RateLimit == 0 {
			flag.IntVar(&sets.RateLimit, "l", 3, "rate limit of worker pool")
		}
	}

	flag.Parse()
}
