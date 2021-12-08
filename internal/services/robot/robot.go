package robot

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/cgriceld/crypto-trade-bot/internal/domain"
)

const (
	FailSendOrderBot = "❌ Fail to place order"
	FailExecOrderBot = "❌ Fail to execute order"
)

var (
	NotSet          = errors.New("Fail to start, parameter wasn't set")
	RunSubscription = errors.New("Fail to start, subscription is already running")
)

type Kraken interface {
	SetMarket(ctx context.Context, m domain.Market)
	Subscribe(ctx context.Context, m domain.Market) (int, error)
	Start(m domain.Market) <-chan domain.CandleSub
	Stop(ctx context.Context, m domain.Market)
	SendOrder(order domain.Order) (*domain.RespOrder, error)
	Accounts(ctx context.Context) (*domain.AccountsResp, error)
}

type Notifications interface {
	Notify(m domain.Market, message string)
}

type Repository interface {
	SaveOrder(rder domain.Order)
	GetOrders(ctx context.Context) ([]domain.Order, error)
	Close()
}

func (r *Robot) Close() {
	r.StopAll(context.Background())
	r.repo.Close()
}

func (r *Robot) isValidStart(ctx context.Context, m domain.Market) error {
	r.muxAll.RLock()
	v, ok := r.trades[m]
	r.muxAll.RUnlock()

	if !ok {
		return fmt.Errorf("%v: market", NotSet)
	}

	v.muxTrade.RLock()
	defer v.muxTrade.RUnlock()

	if v.active {
		return fmt.Errorf("%v: %v", RunSubscription, m)
	}
	if !v.sellActive && !v.buyActive {
		return fmt.Errorf("%v: %v: orders", NotSet, m)
	}

	return nil
}

func (r *Robot) StartMarket(ctx context.Context, m domain.Market) (int, error) {
	if err := r.isValidStart(ctx, m); err != nil {
		return http.StatusBadRequest, err
	}

	r.trades[m].muxTrade.Lock()
	r.trades[m].active = true
	r.trades[m].muxTrade.Unlock()

	status, err := r.kraken.Subscribe(ctx, m)
	if err != nil {
		r.kraken.Stop(ctx, m)
		r.deactivate(m)
		return status, err
	}

	candles := r.kraken.Start(m)
	orders := r.trade(m, candles)

	r.trades[m].wg.Add(1)
	go r.sendOrder(m, orders)

	return 0, nil
}

func (r *Robot) trade(m domain.Market, candles <-chan domain.CandleSub) <-chan domain.Order {
	var ts float64
	orders := make(chan domain.Order)

	r.trades[m].wg.Add(1)
	go func() {
		defer func() {
			close(orders)
			r.trades[m].wg.Done()
		}()

		ts = 0
		for candle := range candles {
			if ts == candle.Cand.Time {
				continue
			}
			ts = candle.Cand.Time
			price, err := r.avgPrice(candle)
			if err != nil {
				r.logger.Errorf("avgPrice: %v: Fail to convert price to float64: %v", m, err)
				continue
			}
			r.logger.Infof("%v: average 1m price: %v", m, price)
			res := r.algo(m, price)
			for _, order := range res {
				orders <- order
			}
		}
	}()

	return orders
}

func (r *Robot) sendOrder(m domain.Market, orders <-chan domain.Order) {
	defer r.trades[m].wg.Done()

	for v := range orders {
		resp, err := r.kraken.SendOrder(v)
		if err != nil {
			r.logger.Errorf("sendOrder: %v: %v: %v", m, v.Typ, err)
			r.notify.Notify(m, fmt.Sprintf("%v: %v: %v: server error", FailSendOrderBot, m, v.Typ))
		}

		r.processOrder(resp, m, v)
	}
}

func (r *Robot) processOrder(respOrder *domain.RespOrder, m domain.Market, v domain.Order) {
	switch {
	// "result":"error"
	case respOrder.Result != "success":
		r.logger.Errorf("processOrder: %v: %v: Fail to send order: %v", m, v.Typ, respOrder.Error)
		r.notify.Notify(m, fmt.Sprintf("%v: %v: %v: server error", FailSendOrderBot, m, v.Typ))

	// balance error
	case respOrder.Status.Stat == "insufficientAvailableFunds":
		r.logger.Warnf("processOrder: %v: %v: Fail to send order: %v", m, v.Typ, respOrder.Status.Stat)
		r.notify.Notify(m, fmt.Sprintf("%v: %v: %v: insufficient funds", FailExecOrderBot, m, v.Typ))

	// order was rejected
	case respOrder.Status.Stat != "placed":
		r.logger.Warnf("processOrder: %v: %v: Fail to send order: %v", m, v.Typ, respOrder.Status.Stat)
		r.notify.Notify(m, fmt.Sprintf("%v: %v: %v", FailExecOrderBot, m, v.Typ))

	// ok
	default:
		r.repo.SaveOrder(v)
		r.logger.Infof(fmt.Sprintf("%s order on %v, price: %.2f", v.Typ, m, v.Price))
		r.notify.Notify(m, fmt.Sprintf("📌 Make %s order on %v. Price: %.2f", v.Typ, m, v.Price))
	}
}

func (r *Robot) deactivate(m domain.Market) {
	r.trades[m].muxTrade.Lock()
	r.trades[m].active = false
	r.trades[m].muxTrade.Unlock()
}

func (r *Robot) StopMarket(ctx context.Context, m domain.Market) error {
	r.muxAll.RLock()
	v, ok := r.trades[m]
	r.muxAll.RUnlock()

	if !ok {
		return fmt.Errorf("%v: %v", NoMarket, m)
	}

	v.muxTrade.RLock()
	status := v.active
	v.muxTrade.RUnlock()

	if !status {
		return nil
	}

	r.kraken.Stop(ctx, m)
	r.trades[m].wg.Wait()
	r.deactivate(m)

	return nil
}

func (r *Robot) StopAll(ctx context.Context) []domain.MarketsResp {
	var res []domain.MarketsResp

	r.muxAll.RLock()
	for m := range r.trades {
		_ = r.StopMarket(ctx, m)
		res = append(res, domain.MarketsResp{
			Market: string(m),
			Status: "ok",
		})
	}
	r.muxAll.RUnlock()

	return res
}

func (r *Robot) StartAll(ctx context.Context) []domain.MarketsResp {
	var res []domain.MarketsResp
	var start domain.MarketsResp

	r.muxAll.RLock()
	for m := range r.trades {
		status, err := r.StartMarket(ctx, m)

		start = domain.MarketsResp{
			Market: string(m),
			Status: "ok",
		}

		if err != nil {
			if status == http.StatusBadRequest {
				start.Status = err.Error()
			} else {
				start.Status = domain.InternalServerError
			}
		}

		res = append(res, start)
	}
	r.muxAll.RUnlock()

	return res
}

func (r *Robot) GetOrders(ctx context.Context) ([]domain.Order, error) {
	return r.repo.GetOrders(ctx)
}
