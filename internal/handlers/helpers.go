package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/cgriceld/crypto-trade-bot/internal/domain"

	"github.com/go-chi/render"
)

var (
	WrongQuery  = errors.New("Wrong query parameter")
	FailedQuery = errors.New("Failed type assertion of query parameter")
)

func (h *Handler) checkPriceSize(w http.ResponseWriter, r *http.Request) (domain.Price, domain.Size) {
	v := r.Context().Value(domain.TriggerPrice)
	if v == nil {
		h.logger.Errorf("%v: %v: no %v", r.URL, WrongQuery, domain.TriggerPrice)
		renderPlain(w, r, http.StatusBadRequest, fmt.Sprintf("%v: no %v", WrongQuery, domain.TriggerPrice))
		return 0, 0
	}
	p, ok := v.(domain.Price)
	if !ok {
		h.logger.Errorf("%v: %v: %v", r.URL, FailedQuery, domain.TriggerPrice)
		renderPlain(w, r, http.StatusInternalServerError, domain.InternalServerError)
		return 0, 0
	}
	if p <= 0 {
		h.logger.Errorf("%v: %v: %v %v", r.URL, WrongQuery, domain.TriggerPrice, p)
		renderPlain(w, r, http.StatusBadRequest, fmt.Sprintf("%v: %v: %v", WrongQuery, domain.TriggerPrice, p))
		return 0, 0
	}

	v = r.Context().Value(domain.OrderSize)
	if v == nil {
		h.logger.Errorf("%v: %v: no %v", r.URL, WrongQuery, domain.OrderSize)
		renderPlain(w, r, http.StatusBadRequest, fmt.Sprintf("%v: no %v", WrongQuery, domain.OrderSize))
		return 0, 0
	}
	s, ok := v.(domain.Size)
	if !ok {
		h.logger.Errorf("%v: %v: %v", r.URL, FailedQuery, domain.OrderSize)
		renderPlain(w, r, http.StatusInternalServerError, domain.InternalServerError)
		return 0, 0
	}
	if s <= 0 {
		h.logger.Errorf("%v: %v: %v %v", r.URL, WrongQuery, domain.OrderSize, s)
		renderPlain(w, r, http.StatusBadRequest, fmt.Sprintf("%v: %v: %v", WrongQuery, domain.OrderSize, s))
		return 0, 0
	}

	return p, s
}

func (h *Handler) checkMarket(w http.ResponseWriter, r *http.Request) domain.Market {
	v := r.Context().Value(domain.MarketName)
	if v == nil {
		h.logger.Errorf("%v: %v: no %v", r.URL, WrongQuery, domain.MarketName)
		renderPlain(w, r, http.StatusBadRequest, fmt.Sprintf("%v: no %v", WrongQuery, domain.MarketName))
		return ""
	}
	m, ok := v.(domain.Market)
	if !ok {
		h.logger.Errorf("%v: %v: %v", r.URL, FailedQuery, domain.MarketName)
		renderPlain(w, r, http.StatusInternalServerError, domain.InternalServerError)
		return ""
	}
	if m == "" {
		h.logger.Errorf("%v: %v: no %v", r.URL, WrongQuery, domain.MarketName)
		renderPlain(w, r, http.StatusBadRequest, fmt.Sprintf("%v: no %v", WrongQuery, domain.MarketName))
		return ""
	}

	return m
}

func renderPlain(w http.ResponseWriter, r *http.Request, code int, text string) {
	render.Status(r, code)
	render.PlainText(w, r, text)
}
