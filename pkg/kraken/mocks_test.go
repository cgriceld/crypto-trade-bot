package kraken

import (
	"tff-go/trade_bot/internal/domain"

	"tff-go/trade_bot/pkg/log"
)

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
