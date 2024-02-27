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
	"github.com/mbiwapa/gophermart.git/internal/app/http-server/handler/api/user/balance"
	"github.com/mbiwapa/gophermart.git/internal/app/http-server/handler/api/user/balance/withdraw"
	"github.com/mbiwapa/gophermart.git/internal/app/http-server/handler/api/user/login"
	"github.com/mbiwapa/gophermart.git/internal/app/http-server/handler/api/user/orders"
	"github.com/mbiwapa/gophermart.git/internal/app/http-server/handler/api/user/register"
	"github.com/mbiwapa/gophermart.git/internal/app/http-server/middleware/authorize"
	"github.com/mbiwapa/gophermart.git/internal/app/http-server/middleware/decompressor"
	mwLogger "github.com/mbiwapa/gophermart.git/internal/app/http-server/middleware/logger"
	"github.com/mbiwapa/gophermart.git/internal/domain/entity"
	"github.com/mbiwapa/gophermart.git/internal/domain/service"
	"github.com/mbiwapa/gophermart.git/internal/infrastructure/postgre"
	"github.com/mbiwapa/gophermart.git/internal/lib/logger"
)

// HTTPServer is a http.Handler that serves HTTP requests.
type HTTPServer struct {
	server         *http.Server
	logger         *logger.Logger
	userService    *service.UserService
	orderService   *service.OrderService
	balanceService *service.BalanceService
	ctx            context.Context
	config         *config.Config
	orderQueue     chan entity.Order
}

// New returns a new HTTPServer.
func New(ctx context.Context, config *config.Config, logger *logger.Logger, orderQueue chan entity.Order) (*HTTPServer, error) {

	server := &HTTPServer{
		server: &http.Server{
			Addr: config.Addr,
			BaseContext: func(_ net.Listener) context.Context {
				return ctx
			},
		},
		logger:     logger,
		ctx:        ctx,
		config:     config,
		orderQueue: orderQueue,
	}
	return server, nil
}

// Run serves HTTP requests.
func (s *HTTPServer) Run() {
	const op = "internal.app.http-server.server.Run"
	log := s.logger.With(s.logger.StringField("op", op))
	go func() {

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

		//FIXME добавить репозиторий заказов
		s.orderService = service.NewOrderService(s.logger, s.orderQueue)

		//FIXME добавить репозиторий баланса
		s.balanceService = service.NewBalanceService(s.logger)

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
//	@host			localhost:8080
func (s *HTTPServer) newRouter() http.Handler {
	docs.SwaggerInfo.Host = s.config.Addr

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(mwLogger.New(s.logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5, "application/json", "text/html"))
	r.Use(decompressor.New(s.logger))
	r.Use(middleware.Heartbeat("/ping"))
	r.Get("/swagger/*", httpSwagger.Handler())

	r.Post("/api/user/register", register.New(s.logger, s.userService))
	r.Post("/api/user/login", login.New(s.logger, s.userService))

	//Only for authenticated users
	r.Group(func(r chi.Router) {
		r.Use(authorize.New(s.logger, s.userService)) //FIXME  почему запускается 2 раза?
		r.Post("/api/user/orders", orders.NewAdder(s.logger, s.orderService, s.userService))
		r.Get("/api/user/orders", orders.NewAllGetter(s.logger, s.orderService, s.userService))
		r.Get("/api/user/balance", balance.New(s.logger, s.balanceService, s.userService))
		r.Post("/api/user/balance/withdraw", withdraw.New(s.logger, s.balanceService, s.userService))
	})

	return r
}
