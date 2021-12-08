package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cgriceld/crypto-trade-bot/internal/handlers"
	"github.com/cgriceld/crypto-trade-bot/internal/repository"
	"github.com/cgriceld/crypto-trade-bot/internal/services/robot"
	"github.com/cgriceld/crypto-trade-bot/pkg/kraken"
	"github.com/cgriceld/crypto-trade-bot/pkg/log"
	pgs "github.com/cgriceld/crypto-trade-bot/pkg/postgres"
	"github.com/cgriceld/crypto-trade-bot/pkg/telegram"

	"github.com/sirupsen/logrus"
)

const (
	serverShutdownTimeout = 5 * time.Second
)

func main() {
	l := logrus.New()
	logger := log.NewLog(l, logrus.DebugLevel, os.Stdout)

	cfg, err := configApp()
	if err != nil {
		logger.Fatalf("Fail to config app: %v", err)
	}

	pool, err := pgs.NewPool(l, cfg.dsn)
	if err != nil {
		return
	}
	defer pool.Close()

	repo := repository.New(pool, logger)
	notify := telegram.New(logger, cfg.TgChatID, cfg.TgBotURL)

	kraken := kraken.New(logger, notify, cfg.APIPublic, cfg.APIPrivate)
	robot := robot.New(kraken, repo, logger, notify)
	handler := handlers.New(robot, logger)

	baseCtx, baseCancel := context.WithCancel(context.Background())
	defer baseCancel()

	server := http.Server{
		Addr:        cfg.port,
		Handler:     handler.Routes(),
		BaseContext: func(net.Listener) context.Context { return baseCtx },
	}

	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	stopChan := make(chan struct{})

	go func() {
		logger.Info("Start server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Fail to start server: %v", err)
		}
	}()

	go func() {
		<-termChan
		logger.Info("Recieve signal, shutting down the server...")

		baseCancel()
		handler.Close()

		ctx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Errorf("Can't shutdown server: %v", err)
		}

		stopChan <- struct{}{}
	}()

	<-stopChan
	logger.Info("Server successfully shutdown")
}
