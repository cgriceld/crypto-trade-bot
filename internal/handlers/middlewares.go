package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/cgriceld/crypto-trade-bot/internal/domain"
)

func getMarket(handler http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		marketQ := r.URL.Query().Get("market")

		ctx := context.WithValue(r.Context(), domain.MarketName, domain.Market(marketQ))
		handler.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(fn)
}

func getPrice(handler http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		priceQ := r.URL.Query().Get("price")
		var price float64
		price = 0
		if val, err := strconv.ParseFloat(priceQ, 64); err == nil {
			price = val
		}

		ctx := context.WithValue(r.Context(), domain.TriggerPrice, domain.Price(price))
		handler.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(fn)
}

func getSize(handler http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		sizeQ := r.URL.Query().Get("size")
		size := 0
		if val, err := strconv.Atoi(sizeQ); err == nil {
			size = val
		}

		ctx := context.WithValue(r.Context(), domain.OrderSize, domain.Size(size))
		handler.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(fn)
}
