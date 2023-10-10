package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/vorotislav/alert-service/internal/settings/server"
	"go.uber.org/zap"
	"os"
	"time"
)

var (
	ErrNotFound = errors.New("metrics not found")
)

type MemStorage struct {
	log            *zap.Logger
	set            *server.Settings
	CounterMetrics map[string]int64   `json:"counter_metrics"`
	GaugeMetrics   map[string]float64 `json:"gauge_metrics"`

	encoder     *json.Encoder
	decoder     *json.Decoder
	file        *os.File
	saveMetrics bool
	async       bool
}

func NewMemStorage(ctx context.Context, log *zap.Logger, set *server.Settings) (*MemStorage, error) {
	var (
		file        *os.File
		err         error
		saveMetrics bool
	)
	if set.FileStoragePath != "" {
		file, err = os.OpenFile(set.FileStoragePath, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			return nil, fmt.Errorf("cannot open file for metrics: %w", err)
		}
		saveMetrics = true
	}

	store := &MemStorage{
		log:         log.With(zap.String("package", "store")),
		set:         set,
		file:        file,
		encoder:     json.NewEncoder(file),
		decoder:     json.NewDecoder(file),
		saveMetrics: saveMetrics,
	}
	store.CounterMetrics = make(map[string]int64)
	store.GaugeMetrics = make(map[string]float64)

	if set.StoreInterval > 0 {
		go store.asyncLoop(ctx, set.StoreInterval)
		store.async = true
	}

	if set.Restore {
		if err := store.readMetrics(); err != nil {
			log.Info("cannot read metrics",
				zap.Error(err))
		}
	}

	return store, nil
}

func (m *MemStorage) Stop(_ context.Context) error {
	m.log.Debug("stopping store...")

	if m.file != nil {
		return m.file.Close()
	}

	return nil
}

func (m *MemStorage) UpdateCounter(name string, value int64) (int64, error) {
	oldValue := m.CounterMetrics[name]
	oldValue += value
	m.CounterMetrics[name] = oldValue

	return oldValue, nil
}

func (m *MemStorage) GetCounterValue(name string) (int64, error) {
	value, ok := m.CounterMetrics[name]
	if !ok {
		return 0, ErrNotFound
	}

	return value, nil
}

func (m *MemStorage) AllCounterMetrics() ([]byte, error) {
	resp, err := json.Marshal(m.CounterMetrics)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (m *MemStorage) UpdateGauge(name string, value float64) (float64, error) {
	m.GaugeMetrics[name] = value

	return value, nil
}

func (m *MemStorage) GetGaugeValue(name string) (float64, error) {
	value, ok := m.GaugeMetrics[name]
	if !ok {
		return 0, ErrNotFound
	}

	return value, nil
}

func (m *MemStorage) AllGaugeMetrics() ([]byte, error) {
	resp, err := json.Marshal(m.GaugeMetrics)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (m *MemStorage) writeMetrics() error {
	//if err := m.file.Truncate(0); err != nil {
	//	m.log.Info("cannot truncate file", zap.Error(err))
	//}
	if err := m.encoder.Encode(m.CounterMetrics); err != nil {
		return fmt.Errorf("cannot write counter metrics: %w", err)
	}
	if err := m.encoder.Encode(m.GaugeMetrics); err != nil {
		return fmt.Errorf("cannot write gauge metrics: %w", err)
	}

	m.file.Seek(0, 0)
	return nil
}

func (m *MemStorage) readMetrics() error {
	if err := m.decoder.Decode(&m.CounterMetrics); err != nil {
		return fmt.Errorf("cannot read counter metrics: %w", err)
	}
	if err := m.decoder.Decode(&m.GaugeMetrics); err != nil {
		return fmt.Errorf("cannot read gauge metrics: %w", err)
	}
	return nil
}

func (m *MemStorage) asyncLoop(ctx context.Context, timeout int) {
	t := time.NewTicker(time.Duration(timeout) * time.Second)

	for {
		select {
		case <-ctx.Done():
			m.log.Info("context is done")
			if err := m.writeMetrics(); err != nil {
				m.log.Info("cannot write metrics", zap.Error(err))
			}
			return
		case <-t.C:
			m.log.Info("write to file")
			if err := m.writeMetrics(); err != nil {
				m.log.Info("cannot write metrics", zap.Error(err))
			}
		}
	}
}
