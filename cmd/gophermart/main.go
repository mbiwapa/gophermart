package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mbiwapa/gophermart.git/config"
	"github.com/mbiwapa/gophermart.git/internal/app/http-server/server"
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

	log.Info("Creating HTTP server...")
	srv, err := server.New(mainCtx, conf, log)
	if err != nil {
		log.Error("Failed to create HTTP server", log.ErrorField(err))
		os.Exit(1)
	}

	srv.Run()
	<-mainCtx.Done()
	time.Sleep(3 * time.Second)
	log.Info("Good bye!")
}
