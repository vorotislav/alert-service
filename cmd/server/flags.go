package main

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

var flagRunAddr string

type Config struct {
	Address string `env:"ADDRESS"`
}

func parseFlag() {
	flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")

	flag.Parse()

	cfg := Config{}
	if err := env.Parse(&cfg); err == nil {
		if cfg.Address != "" {
			flagRunAddr = cfg.Address
		}
	}
}
