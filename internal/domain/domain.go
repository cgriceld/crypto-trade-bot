package domain

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"strconv"
	"time"
)

const (
	MarketName   Market = "market"
	TriggerPrice Market = "price"
	OrderSize    Market = "size"
)

var (
	InternalServerError = "Internal Server Error"
)

type Market string
type Price float64
type Size int

type Order struct {
	Time   *time.Time `json:"time,omitempty"`
	Market string     `json:"market"`
	Typ    string     `json:"type"`
	Price  float64    `json:"price"`
	Size   int        `json:"size"`
}

type SendStatus struct {
	Stat string `json:"status"`
}

type RespOrder struct {
	Result string     `json:"result"`
	Status SendStatus `json:"sendStatus"`
	Error  string     `json:"error"`
}

type Subscribe struct {
	Event    string   `json:"event"`
	Mess     string   `json:"message,omitempty"`
	Feed     string   `json:"feed"`
	Products []string `json:"product_ids"`
}

type Candle struct {
	Close string  `json:"close"`
	Open  string  `json:"open"`
	High  string  `json:"high"`
	Low   string  `json:"low"`
	Time  float64 `json:"time"`
}

type CandleSub struct {
	Cand Candle `json:"candle"`
}

type AccountsResp struct {
	Fi_xbtusd float64 `json:"fi_xbtusd"`
	Fi_bchusd float64 `json:"fi_bchusd"`
	Fi_ethusd float64 `json:"fi_ethusd"`
	Fi_ltcusd float64 `json:"fi_ltcusd"`
	Fi_xrpusd float64 `json:"fi_xrpusd"`
	Fv_xrpxbt float64 `json:"fv_xrpxbt"`
}

type Auxiliary struct {
	Af float64 `json:"af"`
}

type Funds struct {
	Aux Auxiliary `json:"auxiliary"`
}

type Markets struct {
	Fi_xbtusd Funds `json:"fi_xbtusd"`
	Fi_bchusd Funds `json:"fi_bchusd"`
	Fi_ethusd Funds `json:"fi_ethusd"`
	Fi_ltcusd Funds `json:"fi_ltcusd"`
	Fi_xrpusd Funds `json:"fi_xrpusd"`
	Fv_xrpxbt Funds `json:"fv_xrpxbt"`
}

type Wallet struct {
	Result   string  `json:"result"`
	Accounts Markets `json:"accounts"`
	Error    string  `json:"error"`
}

type MarketsResp struct {
	Market string `json:"market"`
	Status string `json:"status"`
}

type TgSend struct {
	Id   int    `json:"chat_id"`
	Text string `json:"text"`
}

type API struct {
	Public  string
	Private string
}

type Urls struct {
	Ws        string
	SendOrder string
	Accounts  string
}

func NewAPI(APIPublic string, APIPrivate string) *API {
	return &API{
		Public:  APIPublic,
		Private: APIPrivate,
	}
}

func NewUrls() *Urls {
	return &Urls{
		Ws:        "wss://demo-futures.kraken.com/ws/v1?chart",
		SendOrder: "https://demo-futures.kraken.com/derivatives/api/v3/sendorder",
		Accounts:  "https://demo-futures.kraken.com/derivatives/api/v3/accounts",
	}
}

func (api *API) Auth(point string, body string) (string, string) {
	nonce := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	in := body + nonce + point
	hash := sha256.Sum256([]byte(in))
	macKey, _ := base64.StdEncoding.DecodeString(api.Private)
	mac := hmac.New(sha512.New, macKey)
	_, _ = mac.Write(hash[:])
	authent := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return nonce, authent
}
