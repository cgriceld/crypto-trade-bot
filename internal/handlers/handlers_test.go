package handlers

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"tff-go/trade_bot/internal/domain"
	"tff-go/trade_bot/internal/services/robot"
	"tff-go/trade_bot/pkg/kraken"
	"tff-go/trade_bot/pkg/log"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	setMarket  = "/setmarket"
	setSell    = "/setsell"
	unsetSell  = "/unsetsell"
	setBuy     = "/setbuy"
	unsetBuy   = "/unsetbuy"
	unsetAll   = "/unsetall"
	active     = "/active"
	activeAll  = "/activeall"
	stopMarket = "/stop"
	stopAll    = "/stopall"
	running    = "/running"
)

var (
	logger  log.Logger
	notify  TgMock
	storage RepMock
	krak    *kraken.Kraken
	rob     Robot
	handler *Handler
)

func setup() {
	l := logrus.New()
	logger = log.NewLog(l, logrus.DebugLevel, ioutil.Discard)
	notify = NewTgMock(logger, 0, "")
	storage = NewRepMock()
	krak = kraken.New(logger, notify, "", "")
	rob = robot.New(krak, storage, logger, notify)
	handler = New(rob, logger)
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

type Test struct {
	name   string
	method string
	url    string
	status int
	query  map[domain.Market]interface{}
	resp   string
}

func TestSetMarket(t *testing.T) {
	tests := []Test{
		{"Right query", http.MethodPost, setMarket, http.StatusCreated,
			map[domain.Market]interface{}{
				domain.MarketName: domain.Market("pi_ethusd")},
			"{\"market\":\"pi_ethusd\",\"status\":\"ok\"}\n"},
		{"No query", http.MethodPost, setMarket, http.StatusBadRequest,
			map[domain.Market]interface{}{
				domain.MarketName: domain.Market("")},
			"Wrong query parameter: no market"},
		{"Wrong type query", http.MethodPost, setMarket, http.StatusInternalServerError,
			map[domain.Market]interface{}{
				domain.MarketName: "pi_ethusd"},
			"Internal Server Error"},
	}

	for _, test := range tests {
		request := httptest.NewRequest(test.method, test.url, nil)

		var ctx context.Context
		for k, v := range test.query {
			if ctx == nil {
				ctx = context.Background()
			}
			ctx = context.WithValue(ctx, k, v)
		}

		response := httptest.NewRecorder()
		handler.setMarket(response, request.WithContext(ctx))
		body := response.Body.String()

		if !assert.Equal(t, test.status, response.Code, "%v: Expect: %v, Got: %v", test.name, test.status, response.Code) ||
			!assert.Equal(t, test.resp, body, "%v: Expect: %v, Got: %v", test.name, test.resp, body) {
			t.Fatal()
		}
	}
}

func TestUnsetSell(t *testing.T) {
	tests := []Test{
		{"Right query", http.MethodPost, unsetSell, http.StatusOK,
			map[domain.Market]interface{}{
				domain.MarketName: domain.Market("pi_ethusd")},
			"{\"market\":\"pi_ethusd\",\"status\":\"ok\"}\n"},
		{"No query", http.MethodPost, unsetSell, http.StatusBadRequest,
			map[domain.Market]interface{}{
				domain.MarketName: domain.Market("")},
			"Wrong query parameter: no market"},
		{"No such market", http.MethodPost, unsetSell, http.StatusBadRequest,
			map[domain.Market]interface{}{
				domain.MarketName: domain.Market("not_set")},
			"{\"market\":\"not_set\",\"status\":\"No market was set: not_set\"}\n"},
	}

	for _, test := range tests {
		request := httptest.NewRequest(test.method, test.url, nil)

		var ctx context.Context
		for k, v := range test.query {
			if ctx == nil {
				ctx = context.Background()
			}
			ctx = context.WithValue(ctx, k, v)
		}

		response := httptest.NewRecorder()
		handler.unsetSell(response, request.WithContext(ctx))
		body := response.Body.String()

		if !assert.Equal(t, test.status, response.Code, "%v: Expect: %v, Got: %v", test.name, test.status, response.Code) ||
			!assert.Equal(t, test.resp, body, "%v: Expect: %v, Got: %v", test.name, test.resp, body) {
			t.Fatal()
		}
	}
}

func TestUnsetBuy(t *testing.T) {
	tests := []Test{
		{"Right query", http.MethodPost, unsetBuy, http.StatusOK,
			map[domain.Market]interface{}{
				domain.MarketName: domain.Market("pi_ethusd")},
			"{\"market\":\"pi_ethusd\",\"status\":\"ok\"}\n"},
		{"No query", http.MethodPost, unsetBuy, http.StatusBadRequest,
			map[domain.Market]interface{}{
				domain.MarketName: domain.Market("")},
			"Wrong query parameter: no market"},
		{"No such market", http.MethodPost, unsetBuy, http.StatusBadRequest,
			map[domain.Market]interface{}{
				domain.MarketName: domain.Market("not_set")},
			"{\"market\":\"not_set\",\"status\":\"No market was set: not_set\"}\n"},
	}

	for _, test := range tests {
		request := httptest.NewRequest(test.method, test.url, nil)

		var ctx context.Context
		for k, v := range test.query {
			if ctx == nil {
				ctx = context.Background()
			}
			ctx = context.WithValue(ctx, k, v)
		}

		response := httptest.NewRecorder()
		handler.unsetBuy(response, request.WithContext(ctx))
		body := response.Body.String()

		if !assert.Equal(t, test.status, response.Code, "%v: Expect: %v, Got: %v", test.name, test.status, response.Code) ||
			!assert.Equal(t, test.resp, body, "%v: Expect: %v, Got: %v", test.name, test.resp, body) {
			t.Fatal()
		}
	}
}

func TestUnsetAll(t *testing.T) {
	tests := []Test{
		{"Right", http.MethodPost, unsetAll, http.StatusOK,
			map[domain.Market]interface{}{},
			"[{\"market\":\"pi_ethusd\",\"status\":\"ok\"}]\n"},
	}

	for _, test := range tests {
		request := httptest.NewRequest(test.method, test.url, nil)

		response := httptest.NewRecorder()
		handler.unsetAll(response, request)
		body := response.Body.String()

		if !assert.Equal(t, test.status, response.Code, "%v: Expect: %v, Got: %v", test.name, test.status, response.Code) ||
			!assert.Equal(t, test.resp, body, "%v: Expect: %v, Got: %v", test.name, test.resp, body) {
			t.Fatal()
		}
	}
}

func TestSetSell(t *testing.T) {
	tests := []Test{
		{"Right query", http.MethodPost, setSell, http.StatusCreated,
			map[domain.Market]interface{}{
				domain.MarketName:   domain.Market("pi_ethusd"),
				domain.TriggerPrice: domain.Price(4000),
				domain.OrderSize:    domain.Size(5)},
			"{\"market\":\"pi_ethusd\",\"type\":\"sell\",\"price\":4000,\"size\":5}\n"},
		{"No market query", http.MethodPost, setSell, http.StatusBadRequest,
			map[domain.Market]interface{}{
				domain.MarketName:   domain.Market(""),
				domain.TriggerPrice: domain.Price(4000),
				domain.OrderSize:    domain.Size(5)},
			"Wrong query parameter: no market"},
		{"No price query", http.MethodPost, setSell, http.StatusBadRequest,
			map[domain.Market]interface{}{
				domain.MarketName:   domain.Market("pi_ethusd"),
				domain.TriggerPrice: domain.Price(0),
				domain.OrderSize:    domain.Size(5)},
			"Wrong query parameter: price: 0"},
		{"Negative price query", http.MethodPost, setSell, http.StatusBadRequest,
			map[domain.Market]interface{}{
				domain.MarketName:   domain.Market("pi_ethusd"),
				domain.TriggerPrice: domain.Price(-42),
				domain.OrderSize:    domain.Size(5)},
			"Wrong query parameter: price: -42"},
		{"Wrong type price query", http.MethodPost, setSell, http.StatusInternalServerError,
			map[domain.Market]interface{}{
				domain.MarketName:   domain.Market("pi_ethusd"),
				domain.TriggerPrice: 42,
				domain.OrderSize:    domain.Size(5)},
			"Internal Server Error"},
		{"No size query", http.MethodPost, setSell, http.StatusBadRequest,
			map[domain.Market]interface{}{
				domain.MarketName:   domain.Market("pi_ethusd"),
				domain.TriggerPrice: domain.Price(4000),
				domain.OrderSize:    domain.Size(0)},
			"Wrong query parameter: size: 0"},
		{"Negative size query", http.MethodPost, setSell, http.StatusBadRequest,
			map[domain.Market]interface{}{
				domain.MarketName:   domain.Market("pi_ethusd"),
				domain.TriggerPrice: domain.Price(4000),
				domain.OrderSize:    domain.Size(-42)},
			"Wrong query parameter: size: -42"},
		{"Wrong type size query", http.MethodPost, setSell, http.StatusInternalServerError,
			map[domain.Market]interface{}{
				domain.MarketName:   domain.Market("pi_ethusd"),
				domain.TriggerPrice: domain.Price(4000),
				domain.OrderSize:    5},
			"Internal Server Error"},
		{"No such market", http.MethodPost, setSell, http.StatusBadRequest,
			map[domain.Market]interface{}{
				domain.MarketName:   domain.Market("not_set"),
				domain.TriggerPrice: domain.Price(4000),
				domain.OrderSize:    domain.Size(5)},
			"{\"market\":\"not_set\",\"status\":\"No market was set: not_set\"}\n"},
	}

	for _, test := range tests {
		request := httptest.NewRequest(test.method, test.url, nil)

		var ctx context.Context
		for k, v := range test.query {
			if ctx == nil {
				ctx = context.Background()
			}
			ctx = context.WithValue(ctx, k, v)
		}

		response := httptest.NewRecorder()
		handler.setSell(response, request.WithContext(ctx))
		body := response.Body.String()

		if !assert.Equal(t, test.status, response.Code, "%v: Expect: %v, Got: %v", test.name, test.status, response.Code) ||
			!assert.Equal(t, test.resp, body, "%v: Expect: %v, Got: %v", test.name, test.resp, body) {
			t.Fatal()
		}
	}
}

func TestSetBuy(t *testing.T) {
	tests := []Test{
		{"Right query", http.MethodPost, setBuy, http.StatusCreated,
			map[domain.Market]interface{}{
				domain.MarketName:   domain.Market("pi_ethusd"),
				domain.TriggerPrice: domain.Price(4000),
				domain.OrderSize:    domain.Size(5)},
			"{\"market\":\"pi_ethusd\",\"type\":\"buy\",\"price\":4000,\"size\":5}\n"},
		{"No market query", http.MethodPost, setBuy, http.StatusBadRequest,
			map[domain.Market]interface{}{
				domain.MarketName:   domain.Market(""),
				domain.TriggerPrice: domain.Price(4000),
				domain.OrderSize:    domain.Size(5)},
			"Wrong query parameter: no market"},
		{"No price query", http.MethodPost, setBuy, http.StatusBadRequest,
			map[domain.Market]interface{}{
				domain.MarketName:   domain.Market("pi_ethusd"),
				domain.TriggerPrice: domain.Price(0),
				domain.OrderSize:    domain.Size(5)},
			"Wrong query parameter: price: 0"},
		{"Negative price query", http.MethodPost, setBuy, http.StatusBadRequest,
			map[domain.Market]interface{}{
				domain.MarketName:   domain.Market("pi_ethusd"),
				domain.TriggerPrice: domain.Price(-42),
				domain.OrderSize:    domain.Size(5)},
			"Wrong query parameter: price: -42"},
		{"Wrong type price query", http.MethodPost, setBuy, http.StatusInternalServerError,
			map[domain.Market]interface{}{
				domain.MarketName:   domain.Market("pi_ethusd"),
				domain.TriggerPrice: 42,
				domain.OrderSize:    domain.Size(5)},
			"Internal Server Error"},
		{"No size query", http.MethodPost, setBuy, http.StatusBadRequest,
			map[domain.Market]interface{}{
				domain.MarketName:   domain.Market("pi_ethusd"),
				domain.TriggerPrice: domain.Price(4000),
				domain.OrderSize:    domain.Size(0)},
			"Wrong query parameter: size: 0"},
		{"Negative size query", http.MethodPost, setBuy, http.StatusBadRequest,
			map[domain.Market]interface{}{
				domain.MarketName:   domain.Market("pi_ethusd"),
				domain.TriggerPrice: domain.Price(4000),
				domain.OrderSize:    domain.Size(-42)},
			"Wrong query parameter: size: -42"},
		{"Wrong type size query", http.MethodPost, setSell, http.StatusInternalServerError,
			map[domain.Market]interface{}{
				domain.MarketName:   domain.Market("pi_ethusd"),
				domain.TriggerPrice: domain.Price(4000),
				domain.OrderSize:    5},
			"Internal Server Error"},
		{"No such market", http.MethodPost, setBuy, http.StatusBadRequest,
			map[domain.Market]interface{}{
				domain.MarketName:   domain.Market("not_set"),
				domain.TriggerPrice: domain.Price(4000),
				domain.OrderSize:    domain.Size(5)},
			"{\"market\":\"not_set\",\"status\":\"No market was set: not_set\"}\n"},
	}

	for _, test := range tests {
		request := httptest.NewRequest(test.method, test.url, nil)

		var ctx context.Context
		for k, v := range test.query {
			if ctx == nil {
				ctx = context.Background()
			}
			ctx = context.WithValue(ctx, k, v)
		}

		response := httptest.NewRecorder()
		handler.setBuy(response, request.WithContext(ctx))
		body := response.Body.String()

		if !assert.Equal(t, test.status, response.Code, "%v: Expect: %v, Got: %v", test.name, test.status, response.Code) ||
			!assert.Equal(t, test.resp, body, "%v: Expect: %v, Got: %v", test.name, test.resp, body) {
			t.Fatal()
		}
	}
}

func TestActive(t *testing.T) {
	tests := []Test{
		{"Right query", http.MethodGet, active, http.StatusOK,
			map[domain.Market]interface{}{
				domain.MarketName: domain.Market("pi_ethusd")},
			"[{\"market\":\"pi_ethusd\",\"type\":\"sell\",\"price\":4000,\"size\":5},{\"market\":\"pi_ethusd\",\"type\":\"buy\",\"price\":4000,\"size\":5}]\n"},
		{"No query", http.MethodGet, active, http.StatusBadRequest,
			map[domain.Market]interface{}{
				domain.MarketName: domain.Market("")},
			"Wrong query parameter: no market"},
		{"No such market", http.MethodGet, active, http.StatusBadRequest,
			map[domain.Market]interface{}{
				domain.MarketName: domain.Market("not_set")},
			"{\"market\":\"not_set\",\"status\":\"No market was set: not_set\"}\n"},
	}

	for _, test := range tests {
		request := httptest.NewRequest(test.method, test.url, nil)

		var ctx context.Context
		for k, v := range test.query {
			if ctx == nil {
				ctx = context.Background()
			}
			ctx = context.WithValue(ctx, k, v)
		}

		response := httptest.NewRecorder()
		handler.active(response, request.WithContext(ctx))
		body := response.Body.String()

		if !assert.Equal(t, test.status, response.Code, "%v: Expect: %v, Got: %v", test.name, test.status, response.Code) ||
			!assert.Equal(t, test.resp, body, "%v: Expect: %v, Got: %v", test.name, test.resp, body) {
			t.Fatal()
		}
	}
}

func TestActiveAll(t *testing.T) {
	tests := []Test{
		{"Right query", http.MethodGet, activeAll, http.StatusOK,
			map[domain.Market]interface{}{
				domain.MarketName: domain.Market("pi_ethusd")},
			"[{\"market\":\"pi_ethusd\",\"type\":\"sell\",\"price\":4000,\"size\":5},{\"market\":\"pi_ethusd\",\"type\":\"buy\",\"price\":4000,\"size\":5}]\n"},
	}

	for _, test := range tests {
		request := httptest.NewRequest(test.method, test.url, nil)

		var ctx context.Context
		for k, v := range test.query {
			if ctx == nil {
				ctx = context.Background()
			}
			ctx = context.WithValue(ctx, k, v)
		}

		response := httptest.NewRecorder()
		handler.activeAll(response, request.WithContext(ctx))
		body := response.Body.String()

		if !assert.Equal(t, test.status, response.Code, "%v: Expect: %v, Got: %v", test.name, test.status, response.Code) ||
			!assert.Equal(t, test.resp, body, "%v: Expect: %v, Got: %v", test.name, test.resp, body) {
			t.Fatal()
		}
	}
}

func TestStopMarket(t *testing.T) {
	tests := []Test{
		{"Right Market", http.MethodPost, stopMarket, http.StatusOK,
			map[domain.Market]interface{}{
				domain.MarketName: domain.Market("pi_ethusd")},
			"{\"market\":\"pi_ethusd\",\"status\":\"ok\"}\n"},
		{"No such market", http.MethodPost, stopMarket, http.StatusBadRequest,
			map[domain.Market]interface{}{
				domain.MarketName: domain.Market("not_set")},
			"{\"market\":\"not_set\",\"status\":\"No market was set: not_set\"}\n"},
	}

	for _, test := range tests {
		request := httptest.NewRequest(test.method, test.url, nil)

		var ctx context.Context
		for k, v := range test.query {
			if ctx == nil {
				ctx = context.Background()
			}
			ctx = context.WithValue(ctx, k, v)
		}

		response := httptest.NewRecorder()
		krak.SetMarket(context.Background(), test.query[domain.MarketName].(domain.Market))
		handler.stopMarket(response, request.WithContext(ctx))
		body := response.Body.String()

		if !assert.Equal(t, test.status, response.Code, "%v: Expect: %v, Got: %v", test.name, test.status, response.Code) ||
			!assert.Equal(t, test.resp, body, "%v: Expect: %v, Got: %v", test.name, test.resp, body) {
			t.Fatal()
		}
	}
}

func TestStopAll(t *testing.T) {
	tests := []Test{
		{"Right", http.MethodPost, stopAll, http.StatusOK,
			map[domain.Market]interface{}{},
			"[{\"market\":\"pi_ethusd\",\"status\":\"ok\"}]\n"},
	}

	for _, test := range tests {
		request := httptest.NewRequest(test.method, test.url, nil)

		response := httptest.NewRecorder()
		handler.stopAll(response, request)
		body := response.Body.String()

		if !assert.Equal(t, test.status, response.Code, "%v: Expect: %v, Got: %v", test.name, test.status, response.Code) ||
			!assert.Equal(t, test.resp, body, "%v: Expect: %v, Got: %v", test.name, test.resp, body) {
			t.Fatal()
		}
	}
}

func TestRunning(t *testing.T) {
	tests := []Test{
		{"Right", http.MethodPost, running, http.StatusOK,
			map[domain.Market]interface{}{}, "null\n"},
	}

	for _, test := range tests {
		request := httptest.NewRequest(test.method, test.url, nil)

		response := httptest.NewRecorder()
		handler.running(response, request)
		body := response.Body.String()

		if !assert.Equal(t, test.status, response.Code, "%v: Expect: %v, Got: %v", test.name, test.status, response.Code) ||
			!assert.Equal(t, test.resp, body, "%v: Expect: %v, Got: %v", test.name, test.resp, body) {
			t.Fatal()
		}
	}
}
