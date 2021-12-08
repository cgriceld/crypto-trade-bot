package queries

import (
	"context"

	"github.com/cgriceld/crypto-trade-bot/internal/domain"
)

const saveOrder = `INSERT INTO orders(ts, market, type, price, size) VALUES ($1, $2, $3, $4, $5)`

func (q *Queries) SaveOrder(order domain.Order) error {
	_, err := q.pool.Exec(context.Background(), saveOrder, order.Time, order.Market, order.Typ, order.Price, order.Size)
	if err != nil {
		return err
	}

	return nil
}

const getOrders = `SELECT ts, market, type, price, size FROM orders`

func (q *Queries) GetOrders(ctx context.Context) ([]domain.Order, error) {
	rows, err := q.pool.Query(ctx, getOrders)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var o domain.Order
		err = rows.Scan(&o.Time, &o.Market, &o.Typ, &o.Price, &o.Size)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}
