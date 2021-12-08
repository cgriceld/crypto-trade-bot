package robot

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/cgriceld/crypto-trade-bot/internal/domain"
	"github.com/cgriceld/crypto-trade-bot/pkg/kraken"
	"github.com/cgriceld/crypto-trade-bot/pkg/log"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var (
	logger  log.Logger
	notify  TgMock
	storage RepMock
	krak    Kraken
	robot   *Robot
)

func setup() {
	l := logrus.New()
	logger = log.NewLog(l, logrus.DebugLevel, ioutil.Discard)
	notify = NewTgMock(logger, 0, "")
	storage = NewRepMock()
	krak = kraken.New(logger, notify, "", "")
	robot = New(krak, storage, logger, notify)
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

type SetMarket struct {
	name   string
	market domain.Market
	res    Trade
}

func TestSetMarket(t *testing.T) {
	tests := []SetMarket{
		{"First Set", "pi_ethusd", Trade{}},
		{"Repeat Set", "pi_ethusd", Trade{}},
	}

	for _, test := range tests {
		robot.SetMarket(context.Background(), test.market)

		if !reflect.DeepEqual(*robot.trades[test.market], test.res) {
			t.Fatalf("%v: Expect: %v, Got: %v", test.name, test.res, *robot.trades[test.market])
		}
	}
}

var (
	candlesSample = []domain.CandleSub{
		{Cand: domain.Candle{
			Time:  42.2,
			Close: "42.2",
			Open:  "42.2",
			High:  "42.2",
			Low:   "42.2",
		}},
		{Cand: domain.Candle{
			Time:  21.1,
			Close: "21.4",
			Open:  "21.4",
			High:  "21.4",
			Low:   "21.4",
		}},
		{Cand: domain.Candle{
			Time:  21.1,
			Close: "24.1",
			Open:  "24.1",
			High:  "24.1",
			Low:   "24.1",
		}},
	}
	timeOrders   = time.Now()
	ordersSample = []domain.Order{
		{
			Time:   &timeOrders,
			Market: "pi_ethusd",
			Typ:    "sell",
			Price:  42.2,
			Size:   1,
		},
		{
			Time:   &timeOrders,
			Market: "pi_ethusd",
			Typ:    "buy",
			Price:  21.4,
			Size:   2,
		},
	}
)

type TradeAlgo struct {
	name   string
	market domain.Market
	prices []float64
	sizes  []int
	orders []domain.Order
}

func TestTradeAlgo(t *testing.T) {
	tests := []TradeAlgo{
		{"Trade", domain.Market("pi_ethusd"), []float64{42.2, 21.4}, []int{1, 2}, ordersSample},
	}

	for _, test := range tests {
		_ = robot.SetSell(context.Background(), test.market, domain.Price(test.prices[0]), domain.Size(test.sizes[0]))
		_ = robot.SetBuy(context.Background(), test.market, domain.Price(test.prices[1]), domain.Size(test.sizes[1]))

		candles := make(chan domain.CandleSub)
		orders := robot.trade(test.market, candles)
		go func() {
			for _, sample := range candlesSample {
				candles <- sample
			}
			close(candles)
		}()

		var res []domain.Order
		for o := range orders {
			o.Time = &timeOrders
			res = append(res, o)
		}

		if !assert.Equal(t, test.orders, res, "%v: Expect: %v, Got: %v", test.name, test.orders, res) {
			t.Fatal()
		}
	}
}

var (
	testResp = []domain.RespOrder{
		{
			Result: "fail",
		},
		{
			Status: domain.SendStatus{
				Stat: "insufficientAvailableFunds",
			},
		},
		{
			Status: domain.SendStatus{
				Stat: "not placed",
			},
		},
		{
			Result: "success",
			Status: domain.SendStatus{
				Stat: "placed",
			},
		},
	}
)

type SendOrders struct {
	name   string
	market domain.Market
	resp   domain.RespOrder
	res    []domain.Order
}

func TestProcessOrder(t *testing.T) {
	tests := []SendOrders{
		{"Fail", domain.Market("pi_ethusd"), testResp[0], nil},
		{"No Balance", domain.Market("pi_ethusd"), testResp[1], nil},
		{"Not Placed", domain.Market("pi_ethusd"), testResp[2], nil},
		{"Placed", domain.Market("pi_ethusd"), testResp[3], []domain.Order{
			{
				Time:   &timeOrders,
				Market: "",
				Typ:    "",
				Price:  0.0,
				Size:   0,
			},
		},
		},
	}

	for _, test := range tests {
		robot.processOrder(&test.resp, test.market, domain.Order{})

		res, _ := robot.repo.GetOrders(context.Background())
		if res != nil {
			res[0].Time = &timeOrders
		}

		if !assert.Equal(t, test.res, res, "%v: Expect: %v, Got: %v", test.name, test.res, res) {
			t.Fatal()
		}
	}
}

type SellOrders struct {
	name   string
	market domain.Market
	price  domain.Price
	size   domain.Size
	res    Sell
}

func TestSetSell(t *testing.T) {
	tests := []SellOrders{
		{"Right Set", domain.Market("pi_ethusd"), domain.Price(42), domain.Size(21),
			Sell{
				sellActive: true, sellPrice: domain.Price(42), sellSize: domain.Size(21),
			}},
		{"No Such Market", domain.Market("wrong"), domain.Price(42), domain.Size(21),
			Sell{
				sellActive: true, sellPrice: domain.Price(42), sellSize: domain.Size(21),
			}},
	}

	s := robot.trades[domain.Market("pi_ethusd")]

	for _, test := range tests {
		_ = robot.SetSell(context.Background(), test.market, test.price, test.size)

		if !assert.Equal(t, test.res, s.Sell) {
			t.Fatalf("%v: Expect: %v, Got: %v", test.name, test.res, s.Sell)
		}
	}
}

type ValidStart struct {
	name   string
	market domain.Market
	res    error
}

func TestIsValidStart(t *testing.T) {
	tests := []ValidStart{
		{"Valid Start", domain.Market("pi_ethusd"), nil},
		{"No Orders", domain.Market("pi_xbtusd"), fmt.Errorf("%v: %v: orders", NotSet, domain.Market("pi_xbtusd"))},
		{"No Such Market", domain.Market("wrong"), fmt.Errorf("%v: market", NotSet)},
	}
	robot.SetMarket(context.Background(), domain.Market("pi_xbtusd"))

	for _, test := range tests {
		err := robot.isValidStart(context.Background(), test.market)

		if !assert.Equal(t, err, test.res, "%v: Expect: %v, Got: %v", test.name, test.res, err) {
			t.Fatal()
		}
	}
}

type StopAll struct {
	name    string
	markets []domain.Market
	res     []domain.MarketsResp
}

func TestStopAll(t *testing.T) {
	tests := []StopAll{
		{"Valid Stop", []domain.Market{domain.Market("pi_ethusd"), domain.Market("pi_xbtusd")},
			[]domain.MarketsResp{
				{Market: "pi_ethusd", Status: "ok"},
				{Market: "pi_xbtusd", Status: "ok"},
			}},
	}

	for _, test := range tests {
		for _, m := range test.markets {
			krak.SetMarket(context.Background(), m)
			robot.trades[m].active = true
		}

		res := robot.StopAll(context.Background())

		if !assert.ElementsMatch(t, test.res, res, "%v: Expect: %v, Got: %v", test.name, test.res, res) {
			t.Fatal()
		}
	}
}
