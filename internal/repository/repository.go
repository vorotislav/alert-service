package repository

import (
	"context"
	"fmt"

	"github.com/vorotislav/alert-service/internal/model"
	"github.com/vorotislav/alert-service/internal/repository/localstorage"
	"github.com/vorotislav/alert-service/internal/repository/postgres"
	"github.com/vorotislav/alert-service/internal/settings/server"

	"go.uber.org/zap"
)

type Repository interface {
	Stop(ctx context.Context) error
	UpdateMetric(ctx context.Context, metric model.Metrics) (model.Metrics, error)
	GetCounterValue(ctx context.Context, name string) (int64, error)
	GetGaugeValue(ctx context.Context, name string) (float64, error)
	AllMetrics(ctx context.Context) ([]byte, error)
	Ping(ctx context.Context) error
	UpdateMetrics(ctx context.Context, metrics []model.Metrics) error
}

func NewRepository(ctx context.Context, log *zap.Logger, set *server.Settings) (Repository, error) {
	var (
		r   Repository
		err error
	)

	if set.DatabaseDSN == "" {
		r, err = localstorage.NewMemStorage(ctx, log, set)
	} else {
		r, err = postgres.NewStorage(ctx, log, set)
	}

	if err != nil {
		return nil, fmt.Errorf("create repository: %w", err)
	}

	return r, nil
}
