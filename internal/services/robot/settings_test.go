package robot

import (
	"context"
	"testing"

	"github.com/cgriceld/crypto-trade-bot/internal/domain"

	"github.com/stretchr/testify/assert"
)

type BuyOrders struct {
	name   string
	market domain.Market
	price  domain.Price
	size   domain.Size
	res    Buy
}

func TestSetBuy(t *testing.T) {
	tests := []BuyOrders{
		{"Right set", domain.Market("pi_ethusd"), domain.Price(42), domain.Size(21),
			Buy{
				buyActive: true, buyPrice: domain.Price(42), buySize: domain.Size(21),
			}},
		{"No such market", domain.Market("wrong"), domain.Price(42), domain.Size(21),
			Buy{
				buyActive: true, buyPrice: domain.Price(42), buySize: domain.Size(21),
			}},
	}

	s := robot.trades[domain.Market("pi_ethusd")]

	for _, test := range tests {
		_ = robot.SetBuy(context.Background(), test.market, test.price, test.size)

		if !assert.Equal(t, test.res, s.Buy) {
			t.Fatalf("%v: Expect: %v, Got: %v", test.name, test.res, s.Buy)
		}
	}
}

func TestGetActiveAll(t *testing.T) {
	res := []domain.Order{
		{
			Market: "pi_ethusd",
			Typ:    "sell",
			Price:  42,
			Size:   21,
		},
		{
			Market: "pi_ethusd",
			Typ:    "buy",
			Price:  42,
			Size:   21,
		},
	}

	active := robot.GetActiveAll(context.Background())

	if !assert.ElementsMatch(t, res, active) {
		t.Fatalf("%v: Expect: %v, Got: %v", "active all", res, active)
	}
}

func TestRunning(t *testing.T) {
	res := []domain.MarketsResp{
		{
			Market: "pi_ethusd",
			Status: "running",
		},
	}

	robot.trades[domain.Market("pi_ethusd")].active = true
	robot.trades[domain.Market("pi_xbtusd")].active = false
	running := robot.Running(context.Background())

	if !assert.ElementsMatch(t, res, running) {
		t.Fatalf("%v: Expect: %v, Got: %v", "running", res, running)
	}
}

func TestUnsetAll(t *testing.T) {
	res := []domain.MarketsResp{
		{
			Market: "pi_ethusd",
			Status: "ok",
		},
		{
			Market: "pi_xbtusd",
			Status: "ok",
		},
	}

	unset := robot.UnsetAll(context.Background())

	if !assert.ElementsMatch(t, res, unset) {
		t.Fatalf("%v: Expect: %v, Got: %v", "unset all", res, unset)
	}
}
