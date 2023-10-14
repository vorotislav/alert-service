package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vorotislav/alert-service/internal/settings/server"
	"go.uber.org/zap"
)

type Repo struct {
	pool *pgxpool.Pool
	log  *zap.Logger
}

func NewRepo(ctx context.Context, log *zap.Logger, set *server.Settings) (*Repo, error) {
	log.Debug(fmt.Sprintf("Database DSN: %s", set.DatabaseDSN))
	pool, err := pgxpool.New(ctx, set.DatabaseDSN)
	if err != nil {
		log.Error("cannot create pool", zap.Error(err))

		return nil, err
	}

	return &Repo{
		pool: pool,
		log:  log.With(zap.String("package", "repository")),
	}, nil
}

func (r *Repo) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

func (r *Repo) Close() {
	r.pool.Close()
}
