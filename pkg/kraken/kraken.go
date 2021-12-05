package kraken

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"tff-go/trade_bot/internal/domain"
	"tff-go/trade_bot/pkg/log"

	"github.com/gorilla/websocket"
)

const (
	sendOrderEndpoint = "/api/v3/sendorder"
	accountsEndpoint  = "/api/v3/accounts"
)

type Notifications interface {
	Notify(m domain.Market, message string)
}

type Connection struct {
	ws *websocket.Conn
	wg sync.WaitGroup
}

type Conns map[domain.Market]*Connection

type Kraken struct {
	notify Notifications
	logger log.Logger
	urls   *domain.Urls
	keys   *domain.API
	client http.Client
	muxAll sync.Mutex
	conns  Conns
}

func New(logger log.Logger, notify Notifications, APIPublic string, APIPrivate string) *Kraken {
	k := &Kraken{
		notify: notify,
		logger: logger,
		conns:  make(Conns),
	}

	k.urls = domain.NewUrls()
	k.keys = domain.NewAPI(APIPublic, APIPrivate)

	k.client = http.Client{
		Timeout: time.Second * 30,
	}

	return k
}

func (k *Kraken) SetMarket(ctx context.Context, m domain.Market) {
	k.muxAll.Lock()
	_, ok := k.conns[m]
	if !ok {
		k.conns[m] = &Connection{}
	}
	k.muxAll.Unlock()
}

func (k *Kraken) makeRequest(method string, url string, endpoint string, query string) ([]byte, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("Fail to create request: %w", err)
	}

	nonce, authent := k.keys.Auth(endpoint, query)
	req.Header.Set("APIKey", k.keys.Public)
	req.Header.Set("Nonce", nonce)
	req.Header.Set("Authent", authent)

	res, err := k.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Fail to send request: %w", err)
	}

	by, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Fail to read response: %w", err)
	}
	defer res.Body.Close()

	return by, nil
}

func (k *Kraken) Stop(ctx context.Context, m domain.Market) {
	if k.conns[m].ws != nil {
		k.conns[m].ws.Close()
	}

	k.conns[m].wg.Wait()
}

func (k *Kraken) Start(m domain.Market) <-chan domain.CandleSub {
	candles, stopChan := k.listenCandles(m)

	k.conns[m].wg.Add(1)
	go k.keepAlive(m, stopChan)

	return candles
}

func (k *Kraken) SendOrder(order domain.Order) (*domain.RespOrder, error) {
	query := fmt.Sprintf("orderType=ioc&symbol=%v&side=%v&size=%v&limitPrice=%v", order.Market, order.Typ, order.Size, order.Price)

	res, err := k.makeRequest(http.MethodPost, k.urls.SendOrder+"?"+query, sendOrderEndpoint, query)
	if err != nil {
		return nil, err
	}

	var respOrder domain.RespOrder
	err = json.Unmarshal(res, &respOrder)
	if err != nil {
		return nil, fmt.Errorf("Fail to decode response: %w", err)
	}

	return &respOrder, nil
}

func (k *Kraken) Accounts(ctx context.Context) (*domain.AccountsResp, error) {
	res, err := k.makeRequest(http.MethodGet, k.urls.Accounts, accountsEndpoint, "")
	if err != nil {
		return nil, fmt.Errorf("Accounts: %w", err)
	}

	var wallet domain.Wallet
	if err = json.Unmarshal(res, &wallet); err != nil {
		return nil, fmt.Errorf("Accounts: Fail to decode response: %w", err)
	}
	if wallet.Result != "success" {
		return nil, fmt.Errorf("Accounts: Unsuccessful response: %s", wallet.Error)
	}

	acc := &domain.AccountsResp{
		Fi_xbtusd: wallet.Accounts.Fi_xbtusd.Aux.Af,
		Fi_bchusd: wallet.Accounts.Fi_bchusd.Aux.Af,
		Fi_ethusd: wallet.Accounts.Fi_ethusd.Aux.Af,
		Fi_ltcusd: wallet.Accounts.Fi_ltcusd.Aux.Af,
		Fi_xrpusd: wallet.Accounts.Fi_xrpusd.Aux.Af,
		Fv_xrpxbt: wallet.Accounts.Fv_xrpxbt.Aux.Af,
	}
	return acc, nil
}
