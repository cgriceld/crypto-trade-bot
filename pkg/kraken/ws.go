package kraken

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"tff-go/trade_bot/internal/domain"

	"github.com/gorilla/websocket"
)

const (
	pingPeriod     = (pongWait * 9) / 10
	pongWait       = 60 * time.Second
	writeWait      = 10 * time.Second
	wsRetryTimeout = 5 * time.Second
	wsRetryTime    = 3
)

const (
	StopListen  = "⚠️ Stop subscription on market"
	StartListen = "✅ Start subscription on market"
)

func (k *Kraken) Subscribe(ctx context.Context, m domain.Market) (int, error) {
	var resp *http.Response
	var err error

	k.SetMarket(ctx, m)
	for i := 0; i < wsRetryTime; i++ {
		k.conns[m].ws, resp, err = websocket.DefaultDialer.Dial(k.urls.Ws, http.Header{})
		if err == nil {
			break
		}
		k.logger.Warnf("Retry to establish websocket connection: %v: %v: %v", m, resp.StatusCode, err)
		time.Sleep(wsRetryTimeout)
	}
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("Fail to establish websocket connection: %v: %v: %v", m, resp.StatusCode, err)
	}

	sub := &domain.Subscribe{
		Event:    "subscribe",
		Feed:     "candles_trade_1m",
		Products: []string{string(m)},
	}

	k.conns[m].ws.SetWriteDeadline(time.Now().Add(writeWait))
	err = k.conns[m].ws.WriteJSON(sub)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("Fail to send subscribtion request: %v: %w", m, err)
	}

	k.conns[m].ws.SetReadDeadline(time.Now().Add(pongWait))
	for i := 0; i < 2; i++ {
		err = k.conns[m].ws.ReadJSON(sub)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("Fail to read subscribtion response: %v: %w", m, err)
		}
	}
	if sub.Event != "subscribed" {
		return http.StatusBadRequest, fmt.Errorf("Fail to subscribe: %v: %s", m, sub.Mess)
	}

	return 0, nil
}

func (k *Kraken) keepAlive(m domain.Market, stopChan <-chan struct{}) {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		k.conns[m].ws.Close()

		k.conns[m].wg.Done()
	}()

	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			k.conns[m].ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := k.conns[m].ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				k.logger.Errorf("keepAlive: %v: Fail to send ping: %v", m, err)
			}
		}
	}
}

func (k *Kraken) listenCandles(m domain.Market) (<-chan domain.CandleSub, <-chan struct{}) {
	candles := make(chan domain.CandleSub)
	stopChan := make(chan struct{})
	var candle domain.CandleSub

	k.conns[m].wg.Add(1)
	go func() {
		defer func() {
			stopChan <- struct{}{}
			close(stopChan)
			close(candles)

			k.notify.Notify(m, fmt.Sprintf("%v: %v", StopListen, m))
			k.conns[m].wg.Done()
		}()

		k.notify.Notify(m, fmt.Sprintf("%v: %v", StartListen, m))

		k.conns[m].ws.SetReadDeadline(time.Now().Add(pongWait))
		k.conns[m].ws.SetPongHandler(func(string) error { k.conns[m].ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })

		for {
			err := k.conns[m].ws.ReadJSON(&candle)
			if err != nil {
				k.logger.Warnf("listenCandles: %v: Stop listening on websocket: %v", m, err)
				if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
					_, err := k.Subscribe(context.Background(), m)
					if err != nil {
						return
					}
					k.logger.Infof("listenCandles: %v: Restore webscoket connection", m)
					continue
				}
				return
			}
			candles <- candle
		}
	}()

	return candles, stopChan
}
