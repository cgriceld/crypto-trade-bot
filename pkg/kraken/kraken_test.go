package kraken

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgriceld/crypto-trade-bot/internal/domain"
	"github.com/cgriceld/crypto-trade-bot/pkg/log"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var (
	logger log.Logger
	notify TgMock
	kraken *Kraken
)

func setup() {
	l := logrus.New()
	logger = log.NewLog(l, logrus.DebugLevel, ioutil.Discard)
	notify = NewTgMock(logger, 0, "")
	kraken = New(logger, notify, "", "")
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

type SetMarket struct {
	name   string
	market domain.Market
	res    Connection
}

func TestSetMarket(t *testing.T) {
	tests := []SetMarket{
		{"First set", "pi_ethusd", Connection{}},
	}

	for _, test := range tests {
		kraken.SetMarket(context.Background(), test.market)

		if !reflect.DeepEqual(*kraken.conns[test.market], test.res) {
			t.Fatalf("%v: Expect: %v, Got: %v", test.name, test.res, *kraken.conns[test.market])
		}
	}
}

var (
	upgrader      = websocket.Upgrader{}
	subMessage    string
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
)

type Subscribed struct {
	name string
	mess string
	res  int
}

func subscribe(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()

	sub := &domain.Subscribe{
		Event: subMessage,
	}

	for i := 0; i < 2; i++ {
		err = c.WriteJSON(sub)
		if err != nil {
			break
		}
	}
}

func TestSubscribe(t *testing.T) {
	tests := []Subscribed{
		{"Subscribed", "subscribed", 0},
		{"Not Subscribed", "error", http.StatusBadRequest},
	}

	s := httptest.NewServer(http.HandlerFunc(subscribe))
	defer s.Close()

	u := "ws" + strings.TrimPrefix(s.URL, "http")
	kraken.urls.Ws = u

	for _, test := range tests {
		subMessage = test.mess

		res, _ := kraken.Subscribe(context.Background(), domain.Market("pi_ethusd"))

		if !assert.Equal(t, res, test.res, "%v: Expect: %v, Got: %v", test.name, test.res, res) {
			t.Fatal()
		}
	}
}

type ListenCandles struct {
	name   string
	market domain.Market
	res    []domain.CandleSub
}

func sendCandles(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()

	for _, v := range candlesSample {
		err = c.WriteJSON(v)
		if err != nil {
			break
		}
	}
}

func TestListenCandles(t *testing.T) {
	tests := []ListenCandles{
		{"Listen", domain.Market("pi_ethusd"), candlesSample},
	}

	s := httptest.NewServer(http.HandlerFunc(sendCandles))
	defer s.Close()

	u := "ws" + strings.TrimPrefix(s.URL, "http")
	kraken.urls.Ws = u

	for _, test := range tests {
		kraken.conns[test.market].ws, _, _ = websocket.DefaultDialer.Dial(kraken.urls.Ws, http.Header{})
		candles, stopChan := kraken.listenCandles(test.market)
		var res []domain.CandleSub

		go func() {
			for candle := range candles {
				res = append(res, candle)
			}
		}()

		for range stopChan {
		}

		if !assert.Equal(t, test.res, res, "%v: Expect: %v, Got: %v", test.name, test.res, res) {
			t.Fatal()
		}
	}
}

func TestMakeRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Hi"))
	}))
	defer ts.Close()

	by, _ := kraken.makeRequest(http.MethodGet, ts.URL, "", "")

	if !assert.Equal(t, by, []byte("Hi"), "%v: Expect: %v, Got: %v", "plain request", by, []byte("Hi")) {
		t.Fatal()
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
	name string
	res  domain.RespOrder
}

func TestSendOrder(t *testing.T) {
	tests := []SendOrders{
		{"Fail", testResp[0]},
		{"No balance", testResp[1]},
		{"Not placed", testResp[2]},
		{"Placed", testResp[3]},
	}

	i := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		coded, _ := json.Marshal(testResp[i])
		_, _ = w.Write(coded)
		i++
	}))
	defer ts.Close()
	kraken.urls.SendOrder = ts.URL + "?"

	for _, test := range tests {
		res, _ := kraken.SendOrder(domain.Order{})

		if !assert.Equal(t, test.res, *res, "%v: Expect: %v, Got: %v", test.name, test.res, *res) {
			t.Fatal()
		}
	}
}

func TestKeepAlive(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()
	}))
	defer s.Close()

	u := "ws" + strings.TrimPrefix(s.URL, "http")
	kraken.urls.Ws = u

	stopChan := make(chan struct{})
	go func() {
		time.Sleep(time.Second * 5)
		stopChan <- struct{}{}
		close(stopChan)
	}()

	market := domain.Market("pi_ethusd")

	kraken.conns[market].wg.Add(1)
	kraken.keepAlive(market, stopChan)
}

func TestAccounts(t *testing.T) {
	send := domain.Wallet{
		Result: "success",
		Accounts: domain.Markets{
			Fi_xbtusd: domain.Funds{
				Aux: domain.Auxiliary{
					Af: 42.0,
				},
			},
			Fi_bchusd: domain.Funds{
				Aux: domain.Auxiliary{
					Af: 42.0,
				},
			},
			Fi_ethusd: domain.Funds{
				Aux: domain.Auxiliary{
					Af: 42.0,
				},
			},
			Fi_ltcusd: domain.Funds{
				Aux: domain.Auxiliary{
					Af: 42.0,
				},
			},
			Fi_xrpusd: domain.Funds{
				Aux: domain.Auxiliary{
					Af: 42.0,
				},
			},
			Fv_xrpxbt: domain.Funds{
				Aux: domain.Auxiliary{
					Af: 42.0,
				},
			},
		},
	}

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		coded, _ := json.Marshal(send)
		_, _ = w.Write(coded)
	}))
	defer s.Close()

	kraken.urls.Accounts = s.URL
	res, _ := kraken.Accounts(context.Background())

	testRes := domain.AccountsResp{
		Fi_xbtusd: 42.0,
		Fi_bchusd: 42.0,
		Fi_ethusd: 42.0,
		Fi_ltcusd: 42.0,
		Fi_xrpusd: 42.0,
		Fv_xrpxbt: 42.0,
	}

	if !assert.Equal(t, testRes, *res, "%v: Expect: %v, Got: %v", "base test", testRes, *res) {
		t.Fatal()
	}
}
