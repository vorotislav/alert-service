package repository

import (
	"context"
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
	if set.DatabaseDSN == "" {
		return localstorage.NewMemStorage(ctx, log, set)
	} else {
		return postgres.NewStorage(ctx, log, set)
	}
}
