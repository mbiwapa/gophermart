package server

import (
	"context"
	"net"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"golang.org/x/sync/errgroup"

	"github.com/mbiwapa/gophermart.git/config"
	"github.com/mbiwapa/gophermart.git/docs"
	"github.com/mbiwapa/gophermart.git/internal/app/http-server/handler/api/user/register"
	mwLogger "github.com/mbiwapa/gophermart.git/internal/app/http-server/middleware/logger"
	"github.com/mbiwapa/gophermart.git/internal/domain/user/service"
	"github.com/mbiwapa/gophermart.git/internal/infrastructure/user/postgre"
	"github.com/mbiwapa/gophermart.git/internal/lib/logger"
)

// HTTPServer is a http.Handler that serves HTTP requests.
type HTTPServer struct {
	server      *http.Server
	logger      *logger.Logger
	userService *service.UserService
	ctx         context.Context
	config      *config.Config
}

// New returns a new HTTPServer.
func New(ctx context.Context, config *config.Config, logger *logger.Logger) (*HTTPServer, error) {

	server := &HTTPServer{
		server: &http.Server{
			Addr: config.Addr,
			BaseContext: func(_ net.Listener) context.Context {
				return ctx
			},
		},
		logger: logger,
		ctx:    ctx,
		config: config,
	}
	return server, nil
}

// Run serves HTTP requests.
func (s *HTTPServer) Run() {

	go func() {
		const op = "internal.app.http-server.server.Run"
		log := s.logger.With(s.logger.StringField("op", op))

		dbpool, err := pgxpool.New(s.ctx, s.config.DB)
		if err != nil {
			log.Error("Failed to connect to database", log.ErrorField(err))
			os.Exit(1)
		}

		userRepository, err := postgre.NewUserRepository(s.ctx, dbpool, s.logger)
		if err != nil {
			log.Error("Failed to create user repository", log.ErrorField(err))
			os.Exit(1)
		}

		userService := service.NewUserService(userRepository, s.logger, s.config.SecretKey)
		s.userService = userService

		s.server.Handler = s.newRouter()

		g, gCtx := errgroup.WithContext(s.ctx)
		g.Go(func() error {
			log.Info("Starting server: ", log.StringField("Addr", s.server.Addr))
			return s.server.ListenAndServe()
		})
		g.Go(func() error {
			<-gCtx.Done()
			log.Info("Database connection closed")
			dbpool.Close()
			log.Info("Shutdown server!")
			return s.server.Shutdown(context.Background())
		})
		if err := g.Wait(); err != nil {
			log.Info("Exit reason: ", log.ErrorField(err))
		}
	}()
}

// newRouter returns a new chi.Router
//
//	@title			Gophermart API
//	@version		1.0
//	@description	This is a Gophermart server.
//	@contact.name	v.max
//	@contact.url	http://v.max.example
//	@contact.email	support@example.com
//	@BasePath		/api
func (s *HTTPServer) newRouter() http.Handler {
	docs.SwaggerInfo.Host = s.config.Addr

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(mwLogger.New(s.logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5, "application/json", "text/html"))
	r.Use(middleware.Heartbeat("/ping"))
	// r.With(auth, handler)
	r.Get("/swagger/*", httpSwagger.Handler())

	r.Post("/api/user/register", register.New(s.logger, s.userService))

	return r
}
