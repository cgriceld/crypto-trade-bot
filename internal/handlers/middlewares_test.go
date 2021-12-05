package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

type Mid struct {
	name  string
	query map[string]string
	resp  string
}

func TestMid(t *testing.T) {
	tests := []Mid{
		{"Right Query", map[string]string{
			"market": "pi_ethusd",
			"price":  "42",
			"size":   "1"},
			"{\"market\":\"pi_ethusd\",\"type\":\"sell\",\"price\":42,\"size\":1}\n"},
	}

	r := chi.NewRouter()
	r.With(getMarket, getPrice, getSize).Post("/", handler.setSell)

	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, test := range tests {
		q := url.Values{}
		for k, v := range test.query {
			q.Add(k, v)
		}
		queryString := q.Encode()

		req := httptest.NewRequest(http.MethodPost, ts.URL+"?"+queryString, nil)
		req.RequestURI = ""

		res, _ := http.DefaultClient.Do(req)

		raw, _ := io.ReadAll(res.Body)
		res.Body.Close()
		body := string(raw)

		if !assert.Equal(t, test.resp, body, "%v: Expect: %v, Got: %v", test.name, test.resp, body) {
			t.Fatal()
		}
	}
}
