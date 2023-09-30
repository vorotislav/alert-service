package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/vorotislav/alert-service/internal/model"
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

const (
	MetricCounter = "counter"
	MetricGauge   = "gauge"
)

func readIMetrics(pollCount int) []Metric[uint64] {
	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)

	return []Metric[uint64]{
		{name: "Alloc", metricType: MetricGauge, value: ms.Alloc},
		{name: "BuckHashSys", metricType: MetricGauge, value: ms.BuckHashSys},
		{name: "Frees", metricType: MetricGauge, value: ms.Frees},
		{name: "GCSys", metricType: MetricGauge, value: ms.GCSys},
		{name: "HeapAlloc", metricType: MetricGauge, value: ms.HeapAlloc},
		{name: "HeapIdle", metricType: MetricGauge, value: ms.HeapIdle},
		{name: "HeapInuse", metricType: MetricGauge, value: ms.HeapInuse},
		{name: "HeapObjects", metricType: MetricGauge, value: ms.HeapObjects},
		{name: "HeapReleased", metricType: MetricGauge, value: ms.HeapReleased},
		{name: "HeapSys", metricType: MetricGauge, value: ms.HeapSys},
		{name: "LastGC", metricType: MetricGauge, value: ms.LastGC},
		{name: "Lookups", metricType: MetricGauge, value: ms.Lookups},
		{name: "MCacheInuse", metricType: MetricGauge, value: ms.MCacheInuse},
		{name: "MCacheSys", metricType: MetricGauge, value: ms.MCacheSys},
		{name: "MSpanInuse", metricType: MetricGauge, value: ms.MSpanInuse},
		{name: "MSpanSys", metricType: MetricGauge, value: ms.MSpanSys},
		{name: "Mallocs", metricType: MetricGauge, value: ms.Mallocs},
		{name: "NextGC", metricType: MetricGauge, value: ms.NextGC},
		{name: "NumForcedGC", metricType: MetricGauge, value: uint64(ms.NumForcedGC)},
		{name: "NumGC", metricType: MetricGauge, value: uint64(ms.NumGC)},
		{name: "OtherSys", metricType: MetricGauge, value: ms.OtherSys},
		{name: "PauseTotalNs", metricType: MetricGauge, value: ms.PauseTotalNs},
		{name: "StackInuse", metricType: MetricGauge, value: ms.StackInuse},
		{name: "StackSys", metricType: MetricGauge, value: ms.StackSys},
		{name: "Sys", metricType: MetricGauge, value: ms.Sys},
		{name: "TotalAlloc", metricType: MetricGauge, value: ms.TotalAlloc},
		{name: "PollCount", metricType: MetricCounter, value: uint64(pollCount)},
	}
}

func readFMetrics() []Metric[float64] {
	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)

	return []Metric[float64]{
		{
			name:       "GCCPUFraction",
			metricType: MetricGauge,
			value:      ms.GCCPUFraction,
		},
		{
			name:       "RandomValue",
			metricType: MetricGauge,
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
	m := model.Metrics{
		ID:    metric.name,
		MType: metric.metricType,
	}

	switch metric.metricType {
	case MetricGauge:
		value := float64(metric.value)
		m.Value = &value
	case MetricCounter:
		value := int64(metric.value)
		m.Delta = &value
	}

	raw, err := json.Marshal(m)
	if err != nil {
		log.Printf("cannot send metric: %s\n", err.Error())

		return
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/update", serverURL),
		"application/json",
		bytes.NewBuffer(raw))

	if err != nil {
		log.Printf("cannot send metric: %s\n", err.Error())
	} else {
		log.Printf("send metric: [%s] value: [%v]\n", metric.name, metric.value)
	}

	if resp != nil {
		resp.Body.Close()
	}
}
