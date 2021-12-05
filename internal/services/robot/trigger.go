package robot

import (
	"strconv"
	"time"

	"tff-go/trade_bot/internal/domain"
)

func (r *Robot) avgPrice(candle domain.CandleSub) (float64, error) {
	close, err := strconv.ParseFloat(candle.Cand.Close, 64)
	if err != nil {
		return 0, err
	}

	open, err := strconv.ParseFloat(candle.Cand.Open, 64)
	if err != nil {
		return 0, err
	}

	high, err := strconv.ParseFloat(candle.Cand.High, 64)
	if err != nil {
		return 0, err
	}

	low, err := strconv.ParseFloat(candle.Cand.Low, 64)
	if err != nil {
		return 0, err
	}

	return (close + open + high + low) / 4, nil
}

func (r *Robot) algo(m domain.Market, v float64) []domain.Order {
	var res []domain.Order

	r.trades[m].muxTrade.Lock()
	ts := time.Now()
	if r.trades[m].sellActive && v >= float64(r.trades[m].sellPrice) {
		res = append(res, domain.Order{
			Time:   &ts,
			Market: string(m),
			Typ:    "sell",
			Price:  v,
			Size:   int(r.trades[m].sellSize),
		})
		r.trades[m].sellActive = false
	}

	if r.trades[m].buyActive && v <= float64(r.trades[m].buyPrice) {
		res = append(res, domain.Order{
			Time:   &ts,
			Market: string(m),
			Typ:    "buy",
			Price:  v,
			Size:   int(r.trades[m].buySize),
		})
		r.trades[m].buyActive = false
	}
	r.trades[m].muxTrade.Unlock()

	return res
}
