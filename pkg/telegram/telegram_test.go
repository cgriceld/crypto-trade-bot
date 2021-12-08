package telegram

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cgriceld/crypto-trade-bot/internal/domain"
	"github.com/cgriceld/crypto-trade-bot/pkg/log"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var (
	logger log.Logger
	tg     *Telegram
)

func setup() {
	l := logrus.New()
	logger = log.NewLog(l, logrus.DebugLevel, ioutil.Discard)
	tg = New(logger, 0, "")
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

type Test struct {
	name string
	text string
}

func TestTelegramOK(t *testing.T) {
	tests := []Test{
		{"OK send", "Hi"},
	}

	var tgMess domain.TgSend
	tgOK := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		by, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		_ = json.Unmarshal(by, &tgMess)
	}))
	defer tgOK.Close()

	tg.url = tgOK.URL

	for _, test := range tests {
		tg.Notify(domain.Market(""), test.text)

		if assert.Equal(t, test.text, tgMess.Text, "%v: Expect: %v, Got: %v", test.name, test.text, tgMess.Text) != true {
			t.Fatal()
		}
	}
}

func TestTelegramFail(t *testing.T) {
	tests := []Test{
		{"not OK send", "Hi"},
	}

	tgFail := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer tgFail.Close()

	tg.url = tgFail.URL

	for _, test := range tests {
		err := tg.sendToBot(test.text)

		if assert.NotEqual(t, err, nil, "%v: Expect: %v, Got: %v", test.name, err, nil) != true {
			t.Fatal()
		}
	}
}
