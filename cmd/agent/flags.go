package main

import "flag"

var (
	flagServerAddr     string
	flagReportInterval int
	flagPollInterval   int
)

func parseFlags() {
	flag.StringVar(&flagServerAddr, "a", "localhost:8080", "server url")
	flag.IntVar(&flagReportInterval, "r", 10, "report interval, sec")
	flag.IntVar(&flagPollInterval, "p", 2, "poll interval, sec")

	flag.Parse()
}
