package storage

import (
	"encoding/json"
	"errors"
)

var (
	ErrNotFound = errors.New("metric not found")
)

type MemStorage struct {
	counterMetrics map[string]int64
	gaugeMetrics   map[string]float64
}

func NewMemStorage() *MemStorage {
	store := &MemStorage{}
	store.counterMetrics = make(map[string]int64)
	store.gaugeMetrics = make(map[string]float64)

	return store
}

func (m *MemStorage) UpdateCounter(name string, value int64) error {
	oldValue := m.counterMetrics[name]
	oldValue += value
	m.counterMetrics[name] = oldValue

	return nil
}

func (m *MemStorage) GetCounterValue(name string) (int64, error) {
	value, ok := m.counterMetrics[name]
	if !ok {
		return 0, ErrNotFound
	}

	return value, nil
}

func (m *MemStorage) AllCounterMetrics() ([]byte, error) {
	resp, err := json.Marshal(m.counterMetrics)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (m *MemStorage) UpdateGauge(name string, value float64) error {
	m.gaugeMetrics[name] = value

	return nil
}

func (m *MemStorage) GetGaugeValue(name string) (float64, error) {
	value, ok := m.gaugeMetrics[name]
	if !ok {
		return 0, ErrNotFound
	}

	return value, nil
}

func (m *MemStorage) AllGaugeMetrics() ([]byte, error) {
	resp, err := json.Marshal(m.gaugeMetrics)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
