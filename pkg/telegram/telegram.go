package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cgriceld/crypto-trade-bot/internal/domain"
	"github.com/cgriceld/crypto-trade-bot/pkg/log"
)

type Telegram struct {
	chatID int
	url    string
	logger log.Logger
	client http.Client
}

func New(logger log.Logger, id int, url string) *Telegram {
	return &Telegram{
		chatID: id,
		url:    url,
		logger: logger,
		client: http.Client{
			Timeout: time.Second * 30,
		},
	}
}

func (tg *Telegram) Notify(m domain.Market, message string) {
	err := tg.sendToBot(message)
	if err != nil {
		tg.logger.Errorf("%v: notify: %v", m, err)
	}
}

func (tg *Telegram) sendToBot(mess string) error {
	send := &domain.TgSend{
		Id:   tg.chatID,
		Text: mess,
	}

	coded, err := json.Marshal(send)
	if err != nil {
		return fmt.Errorf("Fail to marshal message to the bot: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, tg.url, bytes.NewBuffer(coded))
	if err != nil {
		return fmt.Errorf("Fail to create request to the bot: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := tg.client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		return fmt.Errorf("Fail to send message to the bot: %w", err)
	}

	return nil
}
