package handlers

import (
	"context"
	"net/http"
	"tff-go/trade_bot/internal/domain"
	"tff-go/trade_bot/pkg/log"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Robot interface {
	Accounts(ctx context.Context) (*domain.AccountsResp, error)
	GetActive(ctx context.Context, m domain.Market) ([]domain.Order, error)
	GetActiveAll(ctx context.Context) []domain.Order
	SetMarket(ctx context.Context, m domain.Market)
	SetSell(ctx context.Context, m domain.Market, p domain.Price, s domain.Size) error
	UnsetSell(ctx context.Context, m domain.Market) error
	SetBuy(ctx context.Context, m domain.Market, p domain.Price, s domain.Size) error
	UnsetBuy(ctx context.Context, m domain.Market) error
	UnsetAll(ctx context.Context) []domain.MarketsResp
	StartMarket(ctx context.Context, m domain.Market) (int, error)
	StopMarket(ctx context.Context, m domain.Market) error
	StartAll(ctx context.Context) []domain.MarketsResp
	StopAll(ctx context.Context) []domain.MarketsResp
	GetOrders(ctx context.Context) ([]domain.Order, error)
	Running(ctx context.Context) []domain.MarketsResp
	Close()
}

type Handler struct {
	robot  Robot
	logger log.Logger
}

func New(robot Robot, logger log.Logger) *Handler {
	return &Handler{
		robot:  robot,
		logger: logger,
	}
}

func (h *Handler) Close() {
	h.robot.Close()
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.Logger)

	r.Group(func(r chi.Router) {
		r.Get("/accounts", h.accounts)
		r.Get("/orders", h.getOrders)
		r.With(getMarket).Get("/active", h.active)
		r.Get("/activeall", h.activeAll)
		r.Get("/running", h.running)
	})

	r.Group(func(r chi.Router) {
		r.With(getMarket).Post("/setmarket", h.setMarket)
		r.With(getMarket).Post("/unsetsell", h.unsetSell)
		r.With(getMarket).Post("/unsetbuy", h.unsetBuy)
		r.Post("/unsetall", h.unsetAll)
	})

	r.Group(func(r chi.Router) {
		r.Use(getMarket, getPrice, getSize)
		r.Post("/setsell", h.setSell)
		r.Post("/setbuy", h.setBuy)
	})

	r.Group(func(r chi.Router) {
		r.With(getMarket).Post("/start", h.startMarket)
		r.With(getMarket).Post("/stop", h.stopMarket)
		r.Post("/startall", h.startAll)
		r.Post("/stopall", h.stopAll)
	})

	return r
}

func (h *Handler) setMarket(w http.ResponseWriter, r *http.Request) {
	m := h.checkMarket(w, r)
	if m == "" {
		return
	}

	h.robot.SetMarket(r.Context(), m)
	res := &domain.MarketsResp{
		Market: string(m),
		Status: "ok",
	}

	h.logger.Infof("Request to %v succeeded", r.URL)
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, res)
}

func (h *Handler) running(w http.ResponseWriter, r *http.Request) {
	res := h.robot.Running(r.Context())

	h.logger.Infof("Request to %v succeeded", r.URL)
	render.Status(r, http.StatusOK)
	render.JSON(w, r, res)
}

func (h *Handler) getOrders(w http.ResponseWriter, r *http.Request) {
	res, err := h.robot.GetOrders(r.Context())
	if err != nil {
		renderPlain(w, r, http.StatusInternalServerError, domain.InternalServerError)
		return
	}

	h.logger.Infof("Request to %v succeeded", r.URL)
	render.Status(r, http.StatusOK)
	render.JSON(w, r, res)
}

func (h *Handler) accounts(w http.ResponseWriter, r *http.Request) {
	res, err := h.robot.Accounts(r.Context())
	if err != nil {
		h.logger.Errorf("%v: %v", r.URL, err)
		renderPlain(w, r, http.StatusInternalServerError, domain.InternalServerError)
		return
	}

	h.logger.Infof("Request to %v succeeded", r.URL)
	render.Status(r, http.StatusOK)
	render.JSON(w, r, res)
}

func (h *Handler) setSell(w http.ResponseWriter, r *http.Request) {
	p, s := h.checkPriceSize(w, r)
	if p == 0 {
		return
	}
	m := h.checkMarket(w, r)
	if m == "" {
		return
	}

	err := h.robot.SetSell(r.Context(), m, p, s)
	if err != nil {
		res := &domain.MarketsResp{
			Market: string(m),
			Status: err.Error(),
		}

		h.logger.Errorf("%v: %v", r.URL, err)
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, res)
		return
	}

	res := &domain.Order{
		Market: string(m),
		Typ:    "sell",
		Price:  float64(p),
		Size:   int(s),
	}

	h.logger.Infof("Request to %v succeeded", r.URL)
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, res)
}

func (h *Handler) setBuy(w http.ResponseWriter, r *http.Request) {
	p, s := h.checkPriceSize(w, r)
	if p == 0 {
		return
	}
	m := h.checkMarket(w, r)
	if m == "" {
		return
	}

	err := h.robot.SetBuy(r.Context(), m, p, s)
	if err != nil {
		res := &domain.MarketsResp{
			Market: string(m),
			Status: err.Error(),
		}

		h.logger.Errorf("%v: %v", r.URL, err)
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, res)
		return
	}

	res := &domain.Order{
		Market: string(m),
		Typ:    "buy",
		Price:  float64(p),
		Size:   int(s),
	}

	h.logger.Infof("Request to %v succeeded", r.URL)
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, res)
}

func (h *Handler) startMarket(w http.ResponseWriter, r *http.Request) {
	m := h.checkMarket(w, r)
	if m == "" {
		return
	}

	status, err := h.robot.StartMarket(r.Context(), m)
	res := &domain.MarketsResp{
		Market: string(m),
		Status: "ok",
	}

	if err != nil {
		if status == http.StatusBadRequest {
			res.Status = err.Error()
		} else {
			res.Status = domain.InternalServerError
		}

		h.logger.Errorf("%v: %v", r.URL, err)
		render.Status(r, status)
		render.JSON(w, r, res)
		return
	}

	h.logger.Infof("Request to %v succeeded", r.URL)
	render.Status(r, http.StatusOK)
	render.JSON(w, r, res)
}

func (h *Handler) stopMarket(w http.ResponseWriter, r *http.Request) {
	m := h.checkMarket(w, r)
	if m == "" {
		return
	}

	err := h.robot.StopMarket(r.Context(), m)
	res := &domain.MarketsResp{
		Market: string(m),
		Status: "ok",
	}

	if err != nil {
		res.Status = err.Error()

		h.logger.Errorf("%v: %v", r.URL, err)
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, res)
		return
	}

	h.logger.Infof("Request to %v succeeded", r.URL)
	render.Status(r, http.StatusOK)
	render.JSON(w, r, res)
}

func (h *Handler) startAll(w http.ResponseWriter, r *http.Request) {
	res := h.robot.StartAll(r.Context())

	h.logger.Infof("Request to %v succeeded", r.URL)
	render.Status(r, http.StatusOK)
	render.JSON(w, r, res)
}

func (h *Handler) stopAll(w http.ResponseWriter, r *http.Request) {
	res := h.robot.StopAll(r.Context())

	h.logger.Infof("Request to %v succeeded", r.URL)
	render.Status(r, http.StatusOK)
	render.JSON(w, r, res)
}

func (h *Handler) active(w http.ResponseWriter, r *http.Request) {
	m := h.checkMarket(w, r)
	if m == "" {
		return
	}

	res, err := h.robot.GetActive(r.Context(), m)
	if err != nil {
		res := &domain.MarketsResp{
			Market: string(m),
			Status: err.Error(),
		}

		h.logger.Errorf("%v: %v", r.URL, err)
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, res)
		return
	}

	h.logger.Infof("Request to %v succeeded", r.URL)
	render.Status(r, http.StatusOK)
	render.JSON(w, r, res)
}

func (h *Handler) activeAll(w http.ResponseWriter, r *http.Request) {
	res := h.robot.GetActiveAll(r.Context())

	h.logger.Infof("Request to %v succeeded", r.URL)
	render.Status(r, http.StatusOK)
	render.JSON(w, r, res)
}

func (h *Handler) unsetSell(w http.ResponseWriter, r *http.Request) {
	m := h.checkMarket(w, r)
	if m == "" {
		return
	}

	err := h.robot.UnsetSell(r.Context(), m)
	res := &domain.MarketsResp{
		Market: string(m),
		Status: "ok",
	}

	if err != nil {
		res.Status = err.Error()

		h.logger.Errorf("%v: %v", r.URL, err)
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, res)
		return
	}

	h.logger.Infof("Request to %v succeeded", r.URL)
	render.Status(r, http.StatusOK)
	render.JSON(w, r, res)
}

func (h *Handler) unsetBuy(w http.ResponseWriter, r *http.Request) {
	m := h.checkMarket(w, r)
	if m == "" {
		return
	}

	err := h.robot.UnsetBuy(r.Context(), m)
	res := &domain.MarketsResp{
		Market: string(m),
		Status: "ok",
	}

	if err != nil {
		res.Status = err.Error()

		h.logger.Errorf("%v: %v", r.URL, err)
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, res)
		return
	}

	h.logger.Infof("Request to %v succeeded", r.URL)
	render.Status(r, http.StatusOK)
	render.JSON(w, r, res)
}

func (h *Handler) unsetAll(w http.ResponseWriter, r *http.Request) {
	res := h.robot.UnsetAll(r.Context())

	h.logger.Infof("Request to %v succeeded", r.URL)
	render.Status(r, http.StatusOK)
	render.JSON(w, r, res)
}
