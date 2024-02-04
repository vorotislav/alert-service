// Пакет metrics выполняет основную работу агента: сбор метрик по определённому таймауту и отправку на сервер через клиент.
package metrics

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/vorotislav/alert-service/internal/model"
	"github.com/vorotislav/alert-service/internal/settings/agent"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Доступные типы метрик.
const (
	// MetricTypeCounter метрика типа Счётчик.
	MetricTypeCounter = "counter"
	// MetricTypeGauge метрика типа Датчик.
	MetricTypeGauge = "gauge"
)

// Собираемые метрики.
const (
	MetricAlloc           = "Alloc"
	MetricBuckHashSys     = "BuckHashSys"
	MetricFrees           = "Frees"
	MetricGCSys           = "GCSys"
	MetricHeapAlloc       = "HeapAlloc"
	MetricHeapIdle        = "HeapIdle"
	MetricHeapInuse       = "HeapInuse"
	MetricHeapObjects     = "HeapObjects"
	MetricHeapReleased    = "HeapReleased"
	MetricHeapSys         = "HeapSys"
	MetricLastGC          = "LastGC"
	MetricLookups         = "Lookups"
	MetricMCacheInuse     = "MCacheInuse"
	MetricMCacheSys       = "MCacheSys"
	MetricMSpanInuse      = "MSpanInuse"
	MetricMSpanSys        = "MSpanSys"
	MetricMallocs         = "Mallocs"
	MetricNextGC          = "NextGC"
	MetricNumForcedGC     = "NumForcedGC"
	MetricNumGC           = "NumGC"
	MetricOtherSys        = "OtherSys"
	MetricPauseTotalNs    = "PauseTotalNs"
	MetricStackInuse      = "StackInuse"
	MetricStackSys        = "StackSys"
	MetricSys             = "Sys"
	MetricTotalAlloc      = "TotalAlloc"
	MetricPollCount       = "PollCount"
	MetricGCCPUFraction   = "GCCPUFraction"
	MetricRandomValue     = "RandomValue"
	MetricTotalMemory     = "TotalMemory"
	MetricFreeMemory      = "FreeMemory"
	MetricCPUutilization1 = "CPUutilization1"
)

// Client представляет интерфейс для отправки метрик на сервер.
type Client interface {
	SendMetrics(metrics map[string]*model.Metrics) error
}

// Worker основная часть пакета. Содержит в себе логгер, настройки, клиент для отправки метрик, а так же хранит последние метрики.
type Worker struct {
	log    *zap.Logger
	set    *agent.Settings
	client Client
	cancel context.CancelFunc

	pollCount int
	metrics   map[string]*model.Metrics
}

// NewWorker конструктор для Worker.
func NewWorker(log *zap.Logger, set *agent.Settings, client Client) *Worker {
	w := &Worker{
		log:    log.With(zap.String("package", "metrics worker")),
		set:    set,
		client: client,
	}

	w.initMetrics()

	return w
}

// Start метод начинает осуществлять сбор данных в отдельной горутине.
func (w *Worker) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	w.cancel = cancel

	go w.startWorker(ctx)
}

// Stop прерывает выполнение Worker'а и останавливает сбор данных.
func (w *Worker) Stop(_ context.Context) {
	w.cancel()
}

func (w *Worker) startWorker(ctx context.Context) {
	pollTicker := time.NewTicker(time.Duration(w.set.PollInterval) * time.Second)
	reportTicker := time.NewTicker(time.Duration(w.set.ReportInterval) * time.Second)

	eg := &errgroup.Group{}

	for {
		select {
		case <-pollTicker.C:
			w.pollCount++
			w.readMetrics()

			eg.Go(w.readAddMetrics)

			if err := eg.Wait(); err != nil {
				w.log.Error("read additional metrics", zap.Error(err))
			}

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
	}

	return &model.Metrics{
		ID:    name,
		MType: metricType,
		Delta: new(int64),
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
	w.metrics[MetricTotalMemory] = getTemplateMetric(MetricTotalMemory, MetricTypeGauge)
	w.metrics[MetricFreeMemory] = getTemplateMetric(MetricFreeMemory, MetricTypeGauge)
	w.metrics[MetricCPUutilization1] = getTemplateMetric(MetricCPUutilization1, MetricTypeGauge)
}

func (w *Worker) readAddMetrics() error {
	vm, err := mem.VirtualMemory()
	if err != nil {
		w.log.Error("get virtual memory", zap.Error(err))

		return fmt.Errorf("virtual memory: %w", err)
	}

	percent, err := cpu.Percent(0, true)
	if err != nil {
		w.log.Error("get cpu percent", zap.Error(err))

		return fmt.Errorf("cpu percent: %w", err)
	}

	tm := w.metrics[MetricTotalMemory]
	*tm.Value = float64(vm.Total)
	w.metrics[MetricTotalMemory] = tm

	fm := w.metrics[MetricFreeMemory]
	*fm.Value = float64(vm.Free)
	w.metrics[MetricFreeMemory] = fm

	cu := w.metrics[MetricCPUutilization1]
	*cu.Value = percent[0]
	w.metrics[MetricCPUutilization1] = cu

	return nil
}

func (w *Worker) readMetrics() { //nolint:funlen,gocyclo,cyclop
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
			*v.Value = rand.Float64() //nolint:gosec
		}

		w.metrics[k] = v
	}
	w.pollCount++
}
