package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

func main() {

	parseFlags()
	serverURL := "http://" + flagServerAddr

	fmt.Println("server url: ", serverURL)
	fmt.Println("report interval: ", flagReportInterval)
	fmt.Println("poll interval: ", flagPollInterval)

	ms := runtime.MemStats{}
	pollCount := 0
	var randomValue float64
	lastReportTime := time.Now()
	for {
		pollCount += 1
		runtime.ReadMemStats(&ms)
		randomValue = rand.Float64()

		if time.Now().After(lastReportTime.Add(time.Duration(flagReportInterval) * time.Second)) {
			sendMetrics(serverURL, ms, pollCount, randomValue)
			lastReportTime = time.Now()
		}

		time.Sleep(time.Duration(flagPollInterval) * time.Second)
	}

}

func sendMetrics(serverURL string, ms runtime.MemStats, pollCount int, randomValue float64) {
	resp, _ := http.Post(fmt.Sprintf(serverURL+"/update/gauge/Alloc/%d", ms.Alloc), "text/plain", http.NoBody)
	resp.Body.Close()

	resp, _ = http.Post(fmt.Sprintf(serverURL+"/update/gauge/BuckHashSys/%d", ms.BuckHashSys), "text/plain", http.NoBody)
	resp.Body.Close()

	resp, _ = http.Post(fmt.Sprintf(serverURL+"/update/gauge/Frees/%d", ms.Frees), "text/plain", http.NoBody)
	resp.Body.Close()

	resp, _ = http.Post(fmt.Sprintf(serverURL+"/update/gauge/GCCPUFraction/%f", ms.GCCPUFraction), "text/plain", http.NoBody)
	resp.Body.Close()

	resp, _ = http.Post(fmt.Sprintf(serverURL+"/update/gauge/GCSys/%d", ms.GCSys), "text/plain", http.NoBody)
	resp.Body.Close()

	resp, _ = http.Post(fmt.Sprintf(serverURL+"/update/gauge/HeapAlloc/%d", ms.HeapAlloc), "text/plain", http.NoBody)
	resp.Body.Close()

	resp, _ = http.Post(fmt.Sprintf(serverURL+"/update/gauge/HeapIdle/%d", ms.HeapIdle), "text/plain", http.NoBody)
	resp.Body.Close()

	resp, _ = http.Post(fmt.Sprintf(serverURL+"/update/gauge/HeapInuse/%d", ms.HeapInuse), "text/plain", http.NoBody)
	resp.Body.Close()

	resp, _ = http.Post(fmt.Sprintf(serverURL+"/update/gauge/HeapObjects/%d", ms.HeapObjects), "text/plain", http.NoBody)
	resp.Body.Close()

	resp, _ = http.Post(fmt.Sprintf(serverURL+"/update/gauge/HeapReleased/%d", ms.HeapReleased), "text/plain", http.NoBody)
	resp.Body.Close()

	resp, _ = http.Post(fmt.Sprintf(serverURL+"/update/gauge/HeapSys/%d", ms.HeapSys), "text/plain", http.NoBody)
	resp.Body.Close()

	resp, _ = http.Post(fmt.Sprintf(serverURL+"/update/gauge/LastGC/%d", ms.LastGC), "text/plain", http.NoBody)
	resp.Body.Close()

	resp, _ = http.Post(fmt.Sprintf(serverURL+"/update/gauge/Lookups/%d", ms.Lookups), "text/plain", http.NoBody)
	resp.Body.Close()

	resp, _ = http.Post(fmt.Sprintf(serverURL+"/update/gauge/MCacheInuse/%d", ms.MCacheInuse), "text/plain", http.NoBody)
	resp.Body.Close()

	resp, _ = http.Post(fmt.Sprintf(serverURL+"/update/gauge/MCacheSys/%d", ms.MCacheSys), "text/plain", http.NoBody)
	resp.Body.Close()

	resp, _ = http.Post(fmt.Sprintf(serverURL+"/update/gauge/MSpanInuse/%d", ms.MSpanInuse), "text/plain", http.NoBody)
	resp.Body.Close()

	resp, _ = http.Post(fmt.Sprintf(serverURL+"/update/gauge/MSpanSys/%d", ms.MSpanSys), "text/plain", http.NoBody)
	resp.Body.Close()

	resp, _ = http.Post(fmt.Sprintf(serverURL+"/update/gauge/Mallocs/%d", ms.Mallocs), "text/plain", http.NoBody)
	resp.Body.Close()

	resp, _ = http.Post(fmt.Sprintf(serverURL+"/update/gauge/NextGC/%d", ms.NextGC), "text/plain", http.NoBody)
	resp.Body.Close()

	resp, _ = http.Post(fmt.Sprintf(serverURL+"/update/gauge/NumForcedGC/%d", ms.NumForcedGC), "text/plain", http.NoBody)
	resp.Body.Close()

	resp, _ = http.Post(fmt.Sprintf(serverURL+"/update/gauge/NumGC/%d", ms.NumGC), "text/plain", http.NoBody)
	resp.Body.Close()

	resp, _ = http.Post(fmt.Sprintf(serverURL+"/update/gauge/OtherSys/%d", ms.OtherSys), "text/plain", http.NoBody)
	resp.Body.Close()

	resp, _ = http.Post(fmt.Sprintf(serverURL+"/update/gauge/PauseTotalNs/%d", ms.PauseTotalNs), "text/plain", http.NoBody)
	resp.Body.Close()

	resp, _ = http.Post(fmt.Sprintf(serverURL+"/update/gauge/StackInuse/%d", ms.StackInuse), "text/plain", http.NoBody)
	resp.Body.Close()

	resp, _ = http.Post(fmt.Sprintf(serverURL+"/update/gauge/StackSys/%d", ms.StackSys), "text/plain", http.NoBody)
	resp.Body.Close()

	resp, _ = http.Post(fmt.Sprintf(serverURL+"/update/gauge/Sys/%d", ms.Sys), "text/plain", http.NoBody)
	resp.Body.Close()

	resp, _ = http.Post(fmt.Sprintf(serverURL+"/update/gauge/TotalAlloc/%d", ms.TotalAlloc), "text/plain", http.NoBody)
	resp.Body.Close()

	resp, _ = http.Post(fmt.Sprintf(serverURL+"/update/counter/PollCount/%d", pollCount), "text/plain", http.NoBody)
	resp.Body.Close()

	resp, _ = http.Post(fmt.Sprintf(serverURL+"/update/gauge/RandomValue/%f", randomValue), "text/plain", http.NoBody)
	resp.Body.Close()
}
