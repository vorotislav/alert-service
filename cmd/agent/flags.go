package main

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

var (
	flagServerAddr     string
	flagReportInterval int
	flagPollInterval   int
)

type Config struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

func parseFlags() {
	flag.StringVar(&flagServerAddr, "a", "localhost:8080", "server url")
	flag.IntVar(&flagReportInterval, "r", 10, "report interval, sec")
	flag.IntVar(&flagPollInterval, "p", 2, "poll interval, sec")

	flag.Parse()

	cfg := Config{}
	if err := env.Parse(&cfg); err == nil {
		if cfg.Address != "" {
			flagServerAddr = cfg.Address
		}

		if cfg.PollInterval != 0 {
			flagPollInterval = cfg.PollInterval
		}

		if cfg.ReportInterval != 0 {
			flagReportInterval = cfg.ReportInterval
		}
	}

}
