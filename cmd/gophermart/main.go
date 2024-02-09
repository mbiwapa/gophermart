package main

import (
	"database/sql"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/mbiwapa/gophermart.git/internal/app/http-server/server"
	"github.com/mbiwapa/gophermart.git/internal/config"
	"github.com/mbiwapa/gophermart.git/internal/lib/logger"
)

func main() {

	logger := logger.NewLogger()

	config := config.MustLoadConfig()

	db, err := sql.Open("pgx", config.DB)
	defer db.Close()
	if err != nil {
		logger.Error("Failed to connect to database", logger.ErrorField(err))
		os.Exit(1)
	}

	srv, err := server.NewHTTPServer(config, logger, db)
	if err != nil {
		logger.Error("Failed to create HTTP server", logger.ErrorField(err))
		os.Exit(1)
	}

	srv.Run()
}
