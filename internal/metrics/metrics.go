package metrics

import (
	"context"
	"math/rand"
	"runtime"
	"time"

	"github.com/vorotislav/alert-service/internal/model"
	"github.com/vorotislav/alert-service/internal/settings/agent"

	"go.uber.org/zap"
)

const (
	MetricTypeCounter = "counter"
	MetricTypeGauge   = "gauge"
)

const (
	MetricAlloc         = "Alloc"
	MetricBuckHashSys   = "BuckHashSys"
	MetricFrees         = "Frees"
	MetricGCSys         = "GCSys"
	MetricHeapAlloc     = "HeapAlloc"
	MetricHeapIdle      = "HeapIdle"
	MetricHeapInuse     = "HeapInuse"
	MetricHeapObjects   = "HeapObjects"
	MetricHeapReleased  = "HeapReleased"
	MetricHeapSys       = "HeapSys"
	MetricLastGC        = "LastGC"
	MetricLookups       = "Lookups"
	MetricMCacheInuse   = "MCacheInuse"
	MetricMCacheSys     = "MCacheSys"
	MetricMSpanInuse    = "MSpanInuse"
	MetricMSpanSys      = "MSpanSys"
	MetricMallocs       = "Mallocs"
	MetricNextGC        = "NextGC"
	MetricNumForcedGC   = "NumForcedGC"
	MetricNumGC         = "NumGC"
	MetricOtherSys      = "OtherSys"
	MetricPauseTotalNs  = "PauseTotalNs"
	MetricStackInuse    = "StackInuse"
	MetricStackSys      = "StackSys"
	MetricSys           = "Sys"
	MetricTotalAlloc    = "TotalAlloc"
	MetricPollCount     = "PollCount"
	MetricGCCPUFraction = "GCCPUFraction"
	MetricRandomValue   = "RandomValue"
)

type Client interface {
	SendMetrics(metrics map[string]*model.Metrics) error
}

type Worker struct {
	log    *zap.Logger
	set    *agent.Settings
	client Client
	cancel context.CancelFunc

	pollCount int
	metrics   map[string]*model.Metrics
}

func NewWorker(log *zap.Logger, set *agent.Settings, client Client) *Worker {
	w := &Worker{
		log:    log.With(zap.String("package", "metrics worker")),
		set:    set,
		client: client,
	}

	w.initMetrics()

	return w
}

func (w *Worker) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	w.cancel = cancel
	go w.startWorker(ctx)
}

func (w *Worker) Stop(_ context.Context) {
	w.cancel()
}

func (w *Worker) startWorker(ctx context.Context) {
	pollTicker := time.NewTicker(time.Duration(w.set.PollInterval) * time.Second)
	reportTicker := time.NewTicker(time.Duration(w.set.ReportInterval) * time.Second)

	for {
		select {
		case <-pollTicker.C:
			w.pollCount++
			w.readMetrics()
			w.log.Debug("polling metrics", zap.Int("iteration", w.pollCount))
		case <-reportTicker.C:
			w.log.Debug("report metrics")
			if err := w.client.SendMetrics(w.metrics); err != nil {
				w.log.Error("error of report metrics", zap.Error(err))
			}
		case <-ctx.Done():
			w.log.Debug("stop metrics working")
			pollTicker.Stop()
			reportTicker.Stop()
			return
		}
	}
}

func getTemplateMetric(name, metricType string) *model.Metrics {
	if metricType == MetricTypeGauge {
		return &model.Metrics{
			ID:    name,
			MType: metricType,
			Value: new(float64),
		}
	} else {
		return &model.Metrics{
			ID:    name,
			MType: metricType,
			Delta: new(int64),
		}
	}
}

func (w *Worker) initMetrics() {
	w.metrics = make(map[string]*model.Metrics)
	w.metrics[MetricAlloc] = getTemplateMetric(MetricAlloc, MetricTypeGauge)
	w.metrics[MetricBuckHashSys] = getTemplateMetric(MetricBuckHashSys, MetricTypeGauge)
	w.metrics[MetricFrees] = getTemplateMetric(MetricFrees, MetricTypeGauge)
	w.metrics[MetricGCSys] = getTemplateMetric(MetricGCSys, MetricTypeGauge)
	w.metrics[MetricHeapAlloc] = getTemplateMetric(MetricHeapAlloc, MetricTypeGauge)
	w.metrics[MetricHeapIdle] = getTemplateMetric(MetricHeapIdle, MetricTypeGauge)
	w.metrics[MetricHeapInuse] = getTemplateMetric(MetricHeapInuse, MetricTypeGauge)
	w.metrics[MetricHeapObjects] = getTemplateMetric(MetricHeapObjects, MetricTypeGauge)
	w.metrics[MetricHeapReleased] = getTemplateMetric(MetricHeapReleased, MetricTypeGauge)
	w.metrics[MetricHeapSys] = getTemplateMetric(MetricHeapSys, MetricTypeGauge)
	w.metrics[MetricLastGC] = getTemplateMetric(MetricLastGC, MetricTypeGauge)
	w.metrics[MetricLookups] = getTemplateMetric(MetricLookups, MetricTypeGauge)
	w.metrics[MetricMCacheInuse] = getTemplateMetric(MetricMCacheInuse, MetricTypeGauge)
	w.metrics[MetricMCacheSys] = getTemplateMetric(MetricMCacheSys, MetricTypeGauge)
	w.metrics[MetricMSpanInuse] = getTemplateMetric(MetricMSpanInuse, MetricTypeGauge)
	w.metrics[MetricMSpanSys] = getTemplateMetric(MetricMSpanSys, MetricTypeGauge)
	w.metrics[MetricMallocs] = getTemplateMetric(MetricMallocs, MetricTypeGauge)
	w.metrics[MetricNextGC] = getTemplateMetric(MetricNextGC, MetricTypeGauge)
	w.metrics[MetricNumForcedGC] = getTemplateMetric(MetricNumForcedGC, MetricTypeGauge)
	w.metrics[MetricNumGC] = getTemplateMetric(MetricNumGC, MetricTypeGauge)
	w.metrics[MetricOtherSys] = getTemplateMetric(MetricOtherSys, MetricTypeGauge)
	w.metrics[MetricPauseTotalNs] = getTemplateMetric(MetricPauseTotalNs, MetricTypeGauge)
	w.metrics[MetricStackInuse] = getTemplateMetric(MetricStackInuse, MetricTypeGauge)
	w.metrics[MetricStackSys] = getTemplateMetric(MetricStackSys, MetricTypeGauge)
	w.metrics[MetricSys] = getTemplateMetric(MetricSys, MetricTypeGauge)
	w.metrics[MetricTotalAlloc] = getTemplateMetric(MetricTotalAlloc, MetricTypeGauge)
	w.metrics[MetricPollCount] = getTemplateMetric(MetricPollCount, MetricTypeCounter)
	w.metrics[MetricGCCPUFraction] = getTemplateMetric(MetricGCCPUFraction, MetricTypeGauge)
	w.metrics[MetricRandomValue] = getTemplateMetric(MetricRandomValue, MetricTypeGauge)
}

func (w *Worker) readMetrics() {
	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)

	for k, v := range w.metrics {
		k, v := k, v
		switch k {
		case MetricAlloc:
			*v.Value = float64(ms.Alloc)
		case MetricBuckHashSys:
			*v.Value = float64(ms.BuckHashSys)
		case MetricFrees:
			*v.Value = float64(ms.Frees)
		case MetricGCSys:
			*v.Value = float64(ms.GCSys)
		case MetricHeapAlloc:
			*v.Value = float64(ms.HeapAlloc)
		case MetricHeapIdle:
			*v.Value = float64(ms.HeapIdle)
		case MetricHeapInuse:
			*v.Value = float64(ms.HeapInuse)
		case MetricHeapObjects:
			*v.Value = float64(ms.HeapObjects)
		case MetricHeapReleased:
			*v.Value = float64(ms.HeapReleased)
		case MetricHeapSys:
			*v.Value = float64(ms.HeapSys)
		case MetricLastGC:
			*v.Value = float64(ms.LastGC)
		case MetricLookups:
			*v.Value = float64(ms.Lookups)
		case MetricMCacheInuse:
			*v.Value = float64(ms.MCacheInuse)
		case MetricMCacheSys:
			*v.Value = float64(ms.MCacheSys)
		case MetricMSpanInuse:
			*v.Value = float64(ms.MSpanInuse)
		case MetricMSpanSys:
			*v.Value = float64(ms.MSpanSys)
		case MetricMallocs:
			*v.Value = float64(ms.Mallocs)
		case MetricNextGC:
			*v.Value = float64(ms.NextGC)
		case MetricNumForcedGC:
			*v.Value = float64(ms.NumForcedGC)
		case MetricNumGC:
			*v.Value = float64(ms.NumGC)
		case MetricOtherSys:
			*v.Value = float64(ms.OtherSys)
		case MetricPauseTotalNs:
			*v.Value = float64(ms.PauseTotalNs)
		case MetricStackInuse:
			*v.Value = float64(ms.StackInuse)
		case MetricStackSys:
			*v.Value = float64(ms.StackSys)
		case MetricSys:
			*v.Value = float64(ms.Sys)
		case MetricTotalAlloc:
			*v.Value = float64(ms.TotalAlloc)
		case MetricPollCount:
			*v.Delta = int64(w.pollCount)
		case MetricGCCPUFraction:
			*v.Value = ms.GCCPUFraction
		case MetricRandomValue:
			*v.Value = rand.Float64()
		}

		w.metrics[k] = v
	}
	w.pollCount++
}
