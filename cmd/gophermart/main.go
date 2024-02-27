package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mbiwapa/gophermart.git/config"
	"github.com/mbiwapa/gophermart.git/internal/app/http-server/server"
	"github.com/mbiwapa/gophermart.git/internal/app/workers"
	"github.com/mbiwapa/gophermart.git/internal/domain/entity"
	"github.com/mbiwapa/gophermart.git/internal/lib/logger"
)

// run swag init -g internal/app/http-server/server/server.go to generate swagger docs
// run swag fmt -g internal/app/http-server/server/server.go to format swagger docs
func main() {
	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log := logger.NewLogger()

	log.Info("Loading configuration...")
	conf := config.MustLoadConfig()

	log.Info("Create order queue chanel...")
	orderQueue := make(chan entity.Order, 100)
	defer close(orderQueue)

	log.Info("Create balance queue chanel...")
	balanceQueue := make(chan entity.BalanceOperation, 100)
	defer close(balanceQueue)

	log.Info("Create error chanel ...")
	errorChan := make(chan error)
	defer close(errorChan)
	go func() {
		for orderErr := range errorChan {
			log.Error("Error in order worker", log.ErrorField(orderErr))
			os.Exit(1)
		}
	}()

	log.Info("Creating HTTP server...")
	srv, err := server.New(mainCtx, conf, log, orderQueue)
	if err != nil {
		log.Error("Failed to create HTTP server", log.ErrorField(err))
		os.Exit(1)
	}
	srv.Run()

	orderWorker := workers.NewOrderWorker(mainCtx, log, orderQueue, errorChan, balanceQueue)
	orderWorker.Run()

	balanceWorker := workers.NewBalanceWorker(mainCtx, log, balanceQueue, errorChan)
	balanceWorker.Run()

	<-mainCtx.Done()
	time.Sleep(3 * time.Second)
	log.Info("Good bye!")
}
