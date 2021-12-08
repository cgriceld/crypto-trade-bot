package robot

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/cgriceld/crypto-trade-bot/internal/domain"
	"github.com/cgriceld/crypto-trade-bot/pkg/log"
)

var (
	NoMarket = errors.New("No market was set")
)

type Buy struct {
	buyActive bool
	buyPrice  domain.Price
	buySize   domain.Size
}

type Sell struct {
	sellActive bool
	sellPrice  domain.Price
	sellSize   domain.Size
}

type Trade struct {
	Sell
	Buy
	muxTrade sync.RWMutex
	wg       sync.WaitGroup
	active   bool
}

type TradePool map[domain.Market]*Trade

type Robot struct {
	kraken Kraken
	logger log.Logger
	repo   Repository
	notify Notifications
	muxAll sync.RWMutex
	trades TradePool
}

func New(kraken Kraken, repo Repository, logger log.Logger, notify Notifications) *Robot {
	r := &Robot{
		kraken: kraken,
		repo:   repo,
		logger: logger,
		notify: notify,
		trades: make(TradePool),
	}

	return r
}

func (r *Robot) SetMarket(ctx context.Context, m domain.Market) {
	r.muxAll.Lock()
	_, ok := r.trades[m]
	if !ok {
		r.trades[m] = &Trade{}
	}
	r.muxAll.Unlock()
}

func (r *Robot) SetSell(ctx context.Context, m domain.Market, p domain.Price, s domain.Size) error {
	r.muxAll.RLock()
	v, ok := r.trades[m]
	r.muxAll.RUnlock()

	if !ok {
		return fmt.Errorf("%v: %v", NoMarket, m)
	}

	v.muxTrade.Lock()
	v.sellPrice = p
	v.sellSize = s
	v.sellActive = true
	v.muxTrade.Unlock()

	return nil
}

func (r *Robot) UnsetSell(ctx context.Context, m domain.Market) error {
	r.muxAll.RLock()
	v, ok := r.trades[m]
	r.muxAll.RUnlock()

	if !ok {
		return fmt.Errorf("%v: %v", NoMarket, m)
	}

	v.muxTrade.Lock()
	v.sellActive = false
	v.muxTrade.Unlock()

	return nil
}

func (r *Robot) SetBuy(ctx context.Context, m domain.Market, p domain.Price, s domain.Size) error {
	r.muxAll.RLock()
	v, ok := r.trades[m]
	r.muxAll.RUnlock()

	if !ok {
		return fmt.Errorf("%v: %v", NoMarket, m)
	}

	v.muxTrade.Lock()
	v.buyPrice = p
	v.buySize = s
	v.buyActive = true
	v.muxTrade.Unlock()

	return nil
}

func (r *Robot) UnsetBuy(ctx context.Context, m domain.Market) error {
	r.muxAll.RLock()
	v, ok := r.trades[m]
	r.muxAll.RUnlock()

	if !ok {
		return fmt.Errorf("%v: %v", NoMarket, m)
	}

	v.muxTrade.Lock()
	v.buyActive = false
	v.muxTrade.Unlock()

	return nil
}

func (r *Robot) UnsetAll(ctx context.Context) []domain.MarketsResp {
	var res []domain.MarketsResp

	r.muxAll.RLock()
	for m := range r.trades {
		_ = r.UnsetSell(ctx, m)
		_ = r.UnsetBuy(ctx, m)
		res = append(res, domain.MarketsResp{
			Market: string(m),
			Status: "ok",
		})
	}
	r.muxAll.RUnlock()

	return res
}

func (r *Robot) GetActive(ctx context.Context, m domain.Market) ([]domain.Order, error) {
	r.muxAll.RLock()
	status, ok := r.trades[m]
	r.muxAll.RUnlock()

	if !ok {
		return nil, fmt.Errorf("%v: %v", NoMarket, m)
	}

	var res []domain.Order

	status.muxTrade.RLock()
	if status.sellActive {
		res = append(res, domain.Order{
			Market: string(m),
			Typ:    "sell",
			Price:  float64(status.sellPrice),
			Size:   int(status.sellSize),
		})
	}

	if status.buyActive {
		res = append(res, domain.Order{
			Market: string(m),
			Typ:    "buy",
			Price:  float64(status.buyPrice),
			Size:   int(status.buySize),
		})
	}
	status.muxTrade.RUnlock()

	return res, nil
}

func (r *Robot) GetActiveAll(ctx context.Context) []domain.Order {
	var res []domain.Order

	r.muxAll.RLock()
	for m := range r.trades {
		status, _ := r.GetActive(ctx, m)
		res = append(res, status...)
	}
	r.muxAll.RUnlock()

	return res
}

func (r *Robot) Running(ctx context.Context) []domain.MarketsResp {
	var res []domain.MarketsResp
	var run domain.MarketsResp

	r.muxAll.RLock()
	for m, v := range r.trades {
		v.muxTrade.RLock()
		if v.active {
			run = domain.MarketsResp{
				Market: string(m),
				Status: "running",
			}

			res = append(res, run)
		}
		v.muxTrade.RUnlock()
	}
	r.muxAll.RUnlock()

	return res
}

func (r *Robot) Accounts(ctx context.Context) (*domain.AccountsResp, error) {
	return r.kraken.Accounts(ctx)
}
