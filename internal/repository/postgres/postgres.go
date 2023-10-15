package postgres

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/vorotislav/alert-service/internal/model"

	"github.com/vorotislav/alert-service/internal/settings/server"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

//go:embed migrations/*
var migrations embed.FS

var (
	ErrSourceDriver   = errors.New("cannot create source driver")
	ErrSourceInstance = errors.New("cannot create migrate")
	ErrMigrateUp      = errors.New("cannot migrate up")
	ErrCreateStorage  = errors.New("cannot create storage")
)

const (
	MetricTypeCounter = "counter"
	MetricTypeGauge   = "gauge"
)

type Storage struct {
	pool *pgxpool.Pool
	log  *zap.Logger
}

func NewStorage(ctx context.Context, log *zap.Logger, set *server.Settings) (*Storage, error) {
	log.Debug(fmt.Sprintf("Database DSN: %s", set.DatabaseDSN))
	pool, err := pgxpool.New(ctx, set.DatabaseDSN)

	if err != nil {
		log.Error("cannot create pool", zap.Error(err))

		return nil, err
	}

	s := &Storage{
		pool: pool,
		log:  log.With(zap.String("package", "repository")),
	}

	if err := s.migrate(); err != nil {
		log.Error("cannot create tables", zap.Error(err))

		return nil, fmt.Errorf("%w: %w", ErrCreateStorage, err)
	}

	return s, nil
}

func (s *Storage) migrate() error {
	d, err := iofs.New(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("%w:%w", ErrSourceDriver, err)
	}

	connCfg := s.pool.Config().ConnConfig
	m, err := migrate.NewWithSourceInstance("iofs", d,
		fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			connCfg.User, connCfg.Password, connCfg.Host, connCfg.Port, connCfg.Database))
	if err != nil {
		return fmt.Errorf("%w:%w", ErrSourceInstance, err)
	}

	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}

		return fmt.Errorf("%w:%w", ErrMigrateUp, err)
	}

	return nil
}

func (s *Storage) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func (s *Storage) Stop(_ context.Context) error {
	s.pool.Close()

	return nil
}

func (s *Storage) UpdateCounter(ctx context.Context, name string, value int64) (int64, error) {
	delta, err := s.GetCounterValue(ctx, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_, err := s.pool.Exec(ctx, "INSERT INTO metrics (name, type, delta) VALUES ($1, $2, $3)",
				name, MetricTypeCounter, value)
			if err != nil {
				return 0, err
			}

			return value, nil
		}

		return 0, err
	}

	delta += value

	_, err = s.pool.Exec(ctx, "UPDATE metrics SET delta = $1 WHERE name = $2", delta, name)
	if err != nil {
		return 0, err
	}

	return delta, nil
}

func (s *Storage) GetCounterValue(ctx context.Context, name string) (int64, error) {
	var delta int64
	err := s.pool.QueryRow(ctx, "SELECT delta FROM metrics WHERE name = $1", name).Scan(&delta)
	if err != nil {
		return 0, err
	}

	return delta, nil
}

func (s *Storage) AllCounterMetrics(ctx context.Context) ([]byte, error) {
	rows, err := s.pool.Query(ctx, "SELECT name, type, delta FROM metrics WHERE type = $1", MetricTypeCounter)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	counters := make([]model.Metrics, 0)

	for rows.Next() {
		m := model.Metrics{}
		err = rows.Scan(&m.ID, &m.MType, m.Delta)
		if err != nil {
			return nil, err
		}

		counters = append(counters, m)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(counters)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (s *Storage) UpdateGauge(ctx context.Context, name string, value float64) (float64, error) {
	_, err := s.GetGaugeValue(ctx, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_, err = s.pool.Exec(ctx,
				"INSERT INTO metrics (name, type, value) VALUES ($1, $2, $3)", name, MetricTypeGauge, value)
			if err != nil {
				return 0, err
			}
		}

		return 0, err
	}

	_, err = s.pool.Exec(ctx, "UPDATE metrics SET value = $1 WHERE name = $2", value, name)
	if err != nil {
		return 0, err
	}

	return value, nil
}

func (s *Storage) GetGaugeValue(ctx context.Context, name string) (float64, error) {
	var value float64
	err := s.pool.QueryRow(ctx, "SELECT value FROM metrics WHERE name = $1", name).Scan(&value)
	if err != nil {
		return 0, err
	}

	return value, nil
}

func (s *Storage) AllGaugeMetrics(ctx context.Context) ([]byte, error) {
	rows, err := s.pool.Query(ctx, "SELECT name, type, value FROM metrics WHERE type = $1", MetricTypeCounter)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	gauges := make([]model.Metrics, 0)

	for rows.Next() {
		m := model.Metrics{}
		err = rows.Scan(&m.ID, &m.MType, m.Value)
		if err != nil {
			return nil, err
		}

		gauges = append(gauges, m)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(gauges)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
