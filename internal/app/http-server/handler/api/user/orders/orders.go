package orders

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"github.com/mbiwapa/gophermart.git/internal/domain/entity"
	"github.com/mbiwapa/gophermart.git/internal/lib/contexter"
	"github.com/mbiwapa/gophermart.git/internal/lib/logger"
)

// OrderAdder is an interface for adding new order to the user.
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=OrderAdder
type OrderAdder interface {
	Add(ctx context.Context, orderNumber string, userUUID uuid.UUID) error
}

type UserAuthorizer interface {
	Authorize(ctx context.Context, token string) (*entity.User, error)
}

// RequestOrderID request body
type RequestOrderID string

//TODO swag documentation

// NewAdder  returned func for adding a new order to the user.
func NewAdder(log *logger.Logger, adder OrderAdder, authorizer UserAuthorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "app.http-server.handler.api.user.orders.NewAdder"

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		reqID := middleware.GetReqID(ctx)
		ctx = context.WithValue(ctx, contexter.RequestID, reqID)
		logWith := log.With(
			log.StringField("op", op),
			log.StringField("request_id", reqID),
		)

		//FIXME add validation and error handling and etc
		user, err := authorizer.Authorize(ctx, r.Header.Get("Authorization"))
		err = adder.Add(ctx, "123123123", user.UUID)
		if err != nil {
			//FIXME add error handling and etc
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		logWith.Info("User registered")
		w.WriteHeader(http.StatusOK)
	}
}

// AllOrdersGetter is an interface for getting an orders from the user.
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=AllOrderGetter
type AllOrdersGetter interface {
	GetAll(ctx context.Context, userUUID uuid.UUID) ([]entity.Order, error)
}

//TODO swag documentation

// NewAllGetter  returned func for getting all orders from the user.
func NewAllGetter(log *logger.Logger, getter AllOrdersGetter, authorizer UserAuthorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "app.http-server.handler.api.user.orders.NewGetter"

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		reqID := middleware.GetReqID(ctx)
		ctx = context.WithValue(ctx, contexter.RequestID, reqID)
		logWith := log.With(
			log.StringField("op", op),
			log.StringField("request_id", reqID),
		)

		//FIXME add validation and error handling and etc
		user, err := authorizer.Authorize(ctx, r.Header.Get("Authorization"))
		orders, err := getter.GetAll(ctx, user.UUID)
		if err != nil {
			//FIXME add error handling and etc
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		for _, order := range orders {
			//FIXME set order to response
			fmt.Println(order)
		}

		logWith.Info("User registered")
		w.WriteHeader(http.StatusOK)
	}
}
