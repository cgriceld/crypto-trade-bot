package repository

import (
	"context"
	"tff-go/trade_bot/internal/domain"
	"tff-go/trade_bot/internal/repository/queries"
	"tff-go/trade_bot/pkg/log"

	"github.com/jackc/pgx/v4/pgxpool"
)

type repo struct {
	pool   *pgxpool.Pool
	logger log.Logger
	*queries.Queries
}

func New(pool *pgxpool.Pool, logger log.Logger) *repo {
	return &repo{
		pool:    pool,
		logger:  logger,
		Queries: queries.New(pool),
	}
}

func (r *repo) Close() {
	r.pool.Close()
}

func (r *repo) SaveOrder(order domain.Order) {
	_ = r.Queries.SaveOrder(order)
}

func (r *repo) GetOrders(ctx context.Context) ([]domain.Order, error) {
	return r.Queries.GetOrders(ctx)
}
