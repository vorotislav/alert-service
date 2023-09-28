package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

type Metric[T uint64 | float64] struct {
	name       string
	metricType string
	value      T
}

func readIMetrics(pollCount int) []Metric[uint64] {
	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)

	return []Metric[uint64]{
		{name: "Alloc", metricType: "gauge", value: ms.Alloc},
		{name: "BuckHashSys", metricType: "gauge", value: ms.BuckHashSys},
		{name: "Frees", metricType: "gauge", value: ms.Frees},
		{name: "GCSys", metricType: "gauge", value: ms.GCSys},
		{name: "HeapAlloc", metricType: "gauge", value: ms.HeapAlloc},
		{name: "HeapIdle", metricType: "gauge", value: ms.HeapIdle},
		{name: "HeapInuse", metricType: "gauge", value: ms.HeapInuse},
		{name: "HeapObjects", metricType: "gauge", value: ms.HeapObjects},
		{name: "HeapReleased", metricType: "gauge", value: ms.HeapReleased},
		{name: "HeapSys", metricType: "gauge", value: ms.HeapSys},
		{name: "LastGC", metricType: "gauge", value: ms.LastGC},
		{name: "Lookups", metricType: "gauge", value: ms.Lookups},
		{name: "MCacheInuse", metricType: "gauge", value: ms.MCacheInuse},
		{name: "MCacheSys", metricType: "gauge", value: ms.MCacheSys},
		{name: "MSpanInuse", metricType: "gauge", value: ms.MSpanInuse},
		{name: "MSpanSys", metricType: "gauge", value: ms.MSpanSys},
		{name: "Mallocs", metricType: "gauge", value: ms.Mallocs},
		{name: "NextGC", metricType: "gauge", value: ms.NextGC},
		{name: "NumForcedGC", metricType: "gauge", value: uint64(ms.NumForcedGC)},
		{name: "NumGC", metricType: "gauge", value: uint64(ms.NumGC)},
		{name: "OtherSys", metricType: "gauge", value: ms.OtherSys},
		{name: "PauseTotalNs", metricType: "gauge", value: ms.PauseTotalNs},
		{name: "StackInuse", metricType: "gauge", value: ms.StackInuse},
		{name: "StackSys", metricType: "gauge", value: ms.StackSys},
		{name: "Sys", metricType: "gauge", value: ms.Sys},
		{name: "TotalAlloc", metricType: "gauge", value: ms.TotalAlloc},
		{name: "PollCount", metricType: "counter", value: uint64(pollCount)},
	}
}

func readFMetrics() []Metric[float64] {
	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)

	return []Metric[float64]{
		{
			name:       "GCCPUFraction",
			metricType: "gauge",
			value:      ms.GCCPUFraction,
		},
		{
			name:       "RandomValue",
			metricType: "gauge",
			value:      rand.Float64(),
		},
	}
}

func main() {

	parseFlags()
	serverURL := fmt.Sprintf("http://%s", flagServerAddr)

	fmt.Println("server url: ", serverURL)
	fmt.Println("report interval: ", flagReportInterval)
	fmt.Println("poll interval: ", flagPollInterval)

	pollCount := 0
	lastReportTime := time.Now()
	for {
		pollCount += 1

		if time.Now().After(lastReportTime.Add(time.Duration(flagReportInterval) * time.Second)) {
			dMetrics := readIMetrics(pollCount)
			fMetrics := readFMetrics()

			sendMetrics(serverURL, dMetrics)
			sendMetrics(serverURL, fMetrics)
			lastReportTime = time.Now()
		}

		time.Sleep(time.Duration(flagPollInterval) * time.Second)
	}
}

func sendMetrics[T uint64 | float64](serverURL string, metrics []Metric[T]) {
	for _, m := range metrics {
		m := m
		sendMetric(serverURL, m)
	}
}

func sendMetric[T uint64 | float64](serverURL string, metric Metric[T]) {
	resp, err := http.Post(
		fmt.Sprintf("%s/update/%s/%s/%v", serverURL, metric.metricType, metric.name, metric.value),
		"text/plain",
		http.NoBody)

	if err != nil {
		log.Println(fmt.Sprintf("cannot send metric: %s", err.Error()))
	} else {
		log.Println(fmt.Sprintf("send metric: [%s] value: [%v]", metric.name, metric.value))
	}

	if resp != nil {
		resp.Body.Close()
	}
}
