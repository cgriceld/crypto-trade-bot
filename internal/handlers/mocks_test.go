package handlers

import (
	"context"
	"github.com/cgriceld/crypto-trade-bot/internal/domain"

	"github.com/cgriceld/crypto-trade-bot/pkg/log"
)

type RepMock interface {
	SaveOrder(order domain.Order)
	GetOrders(ctx context.Context) ([]domain.Order, error)
	Close()
}

type OrdersInMemory map[string]domain.Order

type ordersStorage struct {
	orders OrdersInMemory
}

func NewRepMock() RepMock {
	return &ordersStorage{
		orders: make(OrdersInMemory),
	}
}

func (s *ordersStorage) SaveOrder(order domain.Order) {
	s.orders[order.Market] = order
}

func (s *ordersStorage) GetOrders(ctx context.Context) ([]domain.Order, error) {
	var res []domain.Order

	for _, v := range s.orders {
		res = append(res, v)
	}

	return res, nil
}

func (s *ordersStorage) Close() {
}

// ============================

type TgMock interface {
	Notify(m domain.Market, message string)
}

type InMemory []string

type messStorage struct {
	logger log.Logger
	mess   InMemory
	id     int
	url    string
}

func NewTgMock(logger log.Logger, id int, url string) *messStorage {
	return &messStorage{
		logger: logger,
		id:     id,
		url:    url,
	}
}

func (tg *messStorage) Notify(m domain.Market, message string) {
	tg.mess = append(tg.mess, message)
}
