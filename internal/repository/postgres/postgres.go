package postgres

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/vorotislav/alert-service/internal/model"
	"github.com/vorotislav/alert-service/internal/settings/server"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
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
	maxRetryAttempt = 4
	retryDelay      = 2
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

func (s *Storage) retryQueryRow(ctx context.Context, query string, result any, args ...any) error {
	delay := time.Second

	var err error

	for i := 1; i <= maxRetryAttempt; i++ {
		s.log.Debug("query row exec",
			zap.String("query", query),
			zap.Any("arg0", args[0]),
			zap.Int("attempt", i))

		err = s.pool.QueryRow(ctx, query, args...).Scan(result)
		if err == nil {
			return nil
		}

		s.log.Debug("cannot exec query row", zap.String("", err.Error()))

		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) {
			return err
		}

		if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
			if i == maxRetryAttempt {
				break
			}

			newDelay := delay + (retryDelay * time.Second)

			s.log.Debug("next attempt in", zap.String("sec", delay.String()))
			time.Sleep(delay)
			delay = newDelay
		}
	}

	return err
}

func (s *Storage) UpdateMetric(ctx context.Context, metric model.Metrics) (model.Metrics, error) {
	var err error
	switch metric.MType {
	case model.MetricGauge:
		{
			var value float64
			err = s.retryQueryRow(ctx,
				`insert into metrics (name, type, value) values ($1, $2, $3) on conflict (name) do update 
					set value = $3 returning value;`,
				&value, metric.ID, metric.MType, metric.Value)
			metric.Value = &value
		}
	case model.MetricCounter:
		{
			var delta int64
			err = s.retryQueryRow(ctx,
				`insert into metrics (name, type, delta) values ($1, $2, $3) on conflict (name) do update 
					set delta = $3 + (select delta from metrics where name = $1) returning delta;`,
				&delta, metric.ID, metric.MType, *metric.Delta)
			metric.Delta = &delta
		}
	}
	if err != nil {
		return model.Metrics{}, fmt.Errorf("update metric: %w", err)
	}

	return metric, nil
}

func (s *Storage) GetCounterValue(ctx context.Context, name string) (int64, error) {
	var delta int64

	err := s.retryQueryRow(ctx, "SELECT delta FROM metrics WHERE name = $1", &delta, name)
	if err != nil {
		return 0, err
	}

	return delta, nil
}

func (s *Storage) AllMetrics(ctx context.Context) ([]byte, error) {
	rows, err := s.pool.Query(ctx, "SELECT name, type, delta, value FROM metrics")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	metrics := make([]model.Metrics, 0)

	for rows.Next() {
		m := model.Metrics{}
		var (
			delta sql.NullInt64
			value sql.NullFloat64
		)
		err = rows.Scan(&m.ID, &m.MType, &delta, &value)
		if err != nil {
			return nil, err
		}

		if delta.Valid {
			m.Delta = &delta.Int64
		}

		if value.Valid {
			m.Value = &value.Float64
		}

		metrics = append(metrics, m)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(metrics)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (s *Storage) GetGaugeValue(ctx context.Context, name string) (float64, error) {
	var value float64

	err := s.retryQueryRow(ctx, "SELECT value FROM metrics WHERE name = $1", &value, name)
	if err != nil {
		return 0, err
	}

	return value, nil
}

func (s *Storage) UpdateMetrics(ctx context.Context, metrics []model.Metrics) error {
	counters := make([]model.Metrics, 0)
	gauges := make([]model.Metrics, 0)
	for _, m := range metrics {
		if m.MType == MetricTypeCounter {
			counters = append(counters, m)
		} else {
			gauges = append(gauges, m)
		}
	}

	if err := s.updateGauges(ctx, gauges); err != nil {
		return fmt.Errorf("cannot update gauge metrics: %w", err)
	}

	if err := s.updateCounters(ctx, counters); err != nil {
		return fmt.Errorf("cannot update counter metrics: %w", err)
	}

	return nil
}

func (s *Storage) updateCounters(ctx context.Context, metrics []model.Metrics) error {
	for _, m := range metrics {
		m := m
		_, err := s.UpdateMetric(ctx, m)
		if err != nil {
			return fmt.Errorf("cannot update counter metric: %w", err)
		}
	}

	return nil
}

func (s *Storage) updateGauges(ctx context.Context, metrics []model.Metrics) error {
	qI := `INSERT INTO metrics (name, type, value) VALUES ($1, $2, $3) on conflict (name) do update 
    set value = $3`

	tx, err := s.pool.Begin(ctx)

	if err != nil {
		return fmt.Errorf("cannot start transaction for update gauge metrics: %w", err)
	}

	defer tx.Rollback(ctx)

	for _, m := range metrics {
		_, err := tx.Exec(ctx, qI, m.ID, m.MType, m.Value)
		if err != nil {
			return fmt.Errorf("cannot update gauge metric: %w", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("cannot commit gauge metrics: %w", err)
	}

	return nil
}
