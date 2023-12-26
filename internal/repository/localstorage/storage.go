package localstorage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/vorotislav/alert-service/internal/model"
	"github.com/vorotislav/alert-service/internal/settings/server"

	"go.uber.org/zap"
)

var (
	ErrNotFound            = errors.New("metrics not found")
	ErrStorageNotAvailable = errors.New("storage not available")
)

type MemStorage struct {
	log     *zap.Logger
	set     *server.Settings
	Metrics map[string]model.Metrics `json:"metrics"`

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
	store.Metrics = make(map[string]model.Metrics)

	if *set.StoreInterval > 0 {
		go store.asyncLoop(ctx, *set.StoreInterval)
		store.async = true
	}

	if *set.Restore {
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

func (m *MemStorage) UpdateMetric(_ context.Context, ms model.Metrics) (model.Metrics, error) {
	metric, ok := m.Metrics[ms.ID]
	if !ok {
		m.Metrics[ms.ID] = ms

		return ms, nil
	}

	switch ms.MType {
	case model.MetricCounter:
		*metric.Delta += *ms.Delta
		m.Metrics[ms.ID] = metric
	default:
		*metric.Value = *ms.Value
		m.Metrics[ms.ID] = metric
	}

	return metric, nil
}

func (m *MemStorage) GetCounterValue(_ context.Context, name string) (int64, error) {
	metric, ok := m.Metrics[name]
	if !ok {
		return 0, ErrNotFound
	}

	return *metric.Delta, nil
}

func (m *MemStorage) AllMetrics(_ context.Context) ([]byte, error) {
	resp, err := json.Marshal(m.Metrics)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (m *MemStorage) GetGaugeValue(_ context.Context, name string) (float64, error) {
	metric, ok := m.Metrics[name]
	if !ok {
		return 0, ErrNotFound
	}

	return *metric.Value, nil
}

func (m *MemStorage) writeMetrics() error {
	//if err := m.file.Truncate(0); err != nil {
	//	m.log.Info("cannot truncate file", zap.Error(err))
	//}
	if err := m.encoder.Encode(m.Metrics); err != nil {
		return fmt.Errorf("cannot write counter metrics: %w", err)
	}

	m.file.Seek(0, 0)
	return nil
}

func (m *MemStorage) readMetrics() error {
	if err := m.decoder.Decode(&m.Metrics); err != nil {
		return fmt.Errorf("cannot read metrics: %w", err)
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

func (m *MemStorage) Ping(_ context.Context) error {
	if m.file == nil {
		return ErrStorageNotAvailable
	}

	return nil
}

func (m *MemStorage) UpdateMetrics(_ context.Context, _ []model.Metrics) error {
	return fmt.Errorf("cannot update metrics in local storage")
}
