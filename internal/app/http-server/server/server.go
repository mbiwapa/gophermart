package server

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/sync/errgroup"

	"github.com/mbiwapa/gophermart.git/internal/app/http-server/handler/api/user/register"
	mwLogger "github.com/mbiwapa/gophermart.git/internal/app/http-server/middleware/logger"
	"github.com/mbiwapa/gophermart.git/internal/config"
	"github.com/mbiwapa/gophermart.git/internal/domain/services"
	"github.com/mbiwapa/gophermart.git/internal/infrastructure/postgre"
	"github.com/mbiwapa/gophermart.git/internal/lib/logger"
)

// HTTPServer is an http.Handler that serves HTTP requests.
type HTTPServer struct {
	server      *http.Server
	logger      *logger.Logger
	userService *services.UserService
}

// NewHTTPServer returns a new HTTPServer.
func NewHTTPServer(config *config.Config, logger *logger.Logger, db *sql.DB) (*HTTPServer, error) {
	const op = "app.http-server.server.NewHTTPServer"
	log := logger.With(logger.StringField("op", op))

	userRepository, err := postgre.NewUserRepository(db, logger)
	if err != nil {
		log.Error("Failed to create user repository", log.ErrorField(err))
		return nil, err
	}
	userService := services.NewUserService(userRepository, logger, config.SecretKey)

	server := &HTTPServer{
		server: &http.Server{
			Addr: config.Addr,
		},
		logger:      logger,
		userService: userService,
	}
	server.server.Handler = server.newRouter()
	return server, nil
}

// Run serves HTTP requests.
func (s *HTTPServer) Run() {
	const op = "internal.app.http-server.server.Run"

	log := s.logger.With(s.logger.StringField("op", op))

	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	s.server.BaseContext = func(_ net.Listener) context.Context {
		return mainCtx
	}

	g, gCtx := errgroup.WithContext(mainCtx)

	g.Go(func() error {
		log.Info("Starting server: ", log.StringField("Addr", s.server.Addr))
		return s.server.ListenAndServe()
	})
	g.Go(func() error {
		<-gCtx.Done()
		log.Info("Shutdown server!")
		return s.server.Shutdown(context.Background())
	})

	if err := g.Wait(); err != nil {
		log.Info("Exit reason: ", log.ErrorField(err))
	}
}

// newRouter returns a new chi.Router
func (s *HTTPServer) newRouter() http.Handler {

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(mwLogger.New(s.logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(middleware.Compress(5, "application/json", "text/html"))
	r.Use(middleware.Heartbeat("/ping"))
	// r.With(auth, handler)

	r.Post("/api/user/register", register.New(s.logger, s.userService))

	return r
}
