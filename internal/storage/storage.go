package storage

type MemStorage struct {
	counterMetrics map[string][]int64
	gaugeMetrics   map[string]float64
}

func NewMemStorage() *MemStorage {
	store := &MemStorage{}
	store.counterMetrics = make(map[string][]int64)
	store.gaugeMetrics = make(map[string]float64)

	return store
}

func (m *MemStorage) UpdateCounter(name string, value int64) error {
	values := m.counterMetrics[name]
	values = append(values, value)
	m.counterMetrics[name] = values

	return nil
}

func (m *MemStorage) UpdateGauge(name string, value float64) error {
	m.gaugeMetrics[name] = value

	return nil
}
