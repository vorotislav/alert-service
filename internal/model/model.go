package model

// Metrics модель для одной метрики.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

const (
	MetricCounter = "counter"
	MetricGauge   = "gauge"
)
