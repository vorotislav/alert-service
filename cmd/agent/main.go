package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

const (
	PollInterval   = 2
	ReportInterval = 10
)

func main() {
	ms := runtime.MemStats{}
	pollCount := 0
	var randomValue float64
	lastReportTime := time.Now()
	for {
		pollCount += 1
		runtime.ReadMemStats(&ms)
		randomValue = rand.Float64()

		if time.Now().After(lastReportTime.Add(ReportInterval * time.Second)) {
			sendMetrics(ms, pollCount, randomValue)
			lastReportTime = time.Now()
		}

		time.Sleep(PollInterval * time.Second)
	}

}

func sendMetrics(ms runtime.MemStats, pollCount int, randomValue float64) {
	http.Post(fmt.Sprintf("http://localhost:8080/update/gauge/Alloc/%d", ms.Alloc), "text/plain", http.NoBody)
	http.Post(fmt.Sprintf("http://localhost:8080/update/gauge/BuckHashSys/%d", ms.BuckHashSys), "text/plain", http.NoBody)
	http.Post(fmt.Sprintf("http://localhost:8080/update/gauge/Frees/%d", ms.Frees), "text/plain", http.NoBody)
	http.Post(fmt.Sprintf("http://localhost:8080/update/gauge/GCCPUFraction/%f", ms.GCCPUFraction), "text/plain", http.NoBody)
	http.Post(fmt.Sprintf("http://localhost:8080/update/gauge/GCSys/%d", ms.GCSys), "text/plain", http.NoBody)
	http.Post(fmt.Sprintf("http://localhost:8080/update/gauge/HeapAlloc/%d", ms.HeapAlloc), "text/plain", http.NoBody)
	http.Post(fmt.Sprintf("http://localhost:8080/update/gauge/HeapIdle/%d", ms.HeapIdle), "text/plain", http.NoBody)
	http.Post(fmt.Sprintf("http://localhost:8080/update/gauge/HeapInuse/%d", ms.HeapInuse), "text/plain", http.NoBody)
	http.Post(fmt.Sprintf("http://localhost:8080/update/gauge/HeapObjects/%d", ms.HeapObjects), "text/plain", http.NoBody)
	http.Post(fmt.Sprintf("http://localhost:8080/update/gauge/HeapReleased/%d", ms.HeapReleased), "text/plain", http.NoBody)
	http.Post(fmt.Sprintf("http://localhost:8080/update/gauge/HeapSys/%d", ms.HeapSys), "text/plain", http.NoBody)
	http.Post(fmt.Sprintf("http://localhost:8080/update/gauge/LastGC/%d", ms.LastGC), "text/plain", http.NoBody)
	http.Post(fmt.Sprintf("http://localhost:8080/update/gauge/Lookups/%d", ms.Lookups), "text/plain", http.NoBody)
	http.Post(fmt.Sprintf("http://localhost:8080/update/gauge/MCacheInuse/%d", ms.MCacheInuse), "text/plain", http.NoBody)
	http.Post(fmt.Sprintf("http://localhost:8080/update/gauge/MCacheSys/%d", ms.MCacheSys), "text/plain", http.NoBody)
	http.Post(fmt.Sprintf("http://localhost:8080/update/gauge/MSpanInuse/%d", ms.MSpanInuse), "text/plain", http.NoBody)
	http.Post(fmt.Sprintf("http://localhost:8080/update/gauge/MSpanSys/%d", ms.MSpanSys), "text/plain", http.NoBody)
	http.Post(fmt.Sprintf("http://localhost:8080/update/gauge/Mallocs/%d", ms.Mallocs), "text/plain", http.NoBody)
	http.Post(fmt.Sprintf("http://localhost:8080/update/gauge/NextGC/%d", ms.NextGC), "text/plain", http.NoBody)
	http.Post(fmt.Sprintf("http://localhost:8080/update/gauge/NumForcedGC/%d", ms.NumForcedGC), "text/plain", http.NoBody)
	http.Post(fmt.Sprintf("http://localhost:8080/update/gauge/NumGC/%d", ms.NumGC), "text/plain", http.NoBody)
	http.Post(fmt.Sprintf("http://localhost:8080/update/gauge/OtherSys/%d", ms.OtherSys), "text/plain", http.NoBody)
	http.Post(fmt.Sprintf("http://localhost:8080/update/gauge/PauseTotalNs/%d", ms.PauseTotalNs), "text/plain", http.NoBody)
	http.Post(fmt.Sprintf("http://localhost:8080/update/gauge/StackInuse/%d", ms.StackInuse), "text/plain", http.NoBody)
	http.Post(fmt.Sprintf("http://localhost:8080/update/gauge/StackSys/%d", ms.StackSys), "text/plain", http.NoBody)
	http.Post(fmt.Sprintf("http://localhost:8080/update/gauge/Sys/%d", ms.Sys), "text/plain", http.NoBody)
	http.Post(fmt.Sprintf("http://localhost:8080/update/gauge/TotalAlloc/%d", ms.TotalAlloc), "text/plain", http.NoBody)
	http.Post(fmt.Sprintf("http://localhost:8080/update/counter/PollCount/%d", pollCount), "text/plain", http.NoBody)
	http.Post(fmt.Sprintf("http://localhost:8080/update/gauge/RandomValue/%f", randomValue), "text/plain", http.NoBody)
}
