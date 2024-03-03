package orders

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"

	"github.com/mbiwapa/gophermart.git/internal/domain/entity"
	"github.com/mbiwapa/gophermart.git/internal/lib/contexter"
	"github.com/mbiwapa/gophermart.git/internal/lib/logger"
	"github.com/mbiwapa/gophermart.git/internal/lib/luna"
)

// OrderAdder is an interface for adding new order to the user.
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=OrderAdder
type OrderAdder interface {
	Add(ctx context.Context, orderNumber int, userUUID uuid.UUID) error
}

// UserAuthorizer is an interface for authorizing users.
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=UserAuthorizer
type UserAuthorizer interface {
	Authorize(ctx context.Context, token string) (*entity.User, error)
}

//TODO swag documentation

// NewAdder  returned func for adding a new order to the user.
func NewAdder(log *logger.Logger, adder OrderAdder, authorizer UserAuthorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "app.http-server.handler.api.user.orders.NewAdder"

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		reqID := middleware.GetReqID(ctx)
		ctx = context.WithValue(ctx, contexter.RequestID, reqID)
		logWith := log.With(
			log.StringField("op", op),
			log.StringField("request_id", reqID),
		)

		user, err := authorizer.Authorize(ctx, r.Header.Get("Authorization"))
		if err != nil {
			logWith.Error("Failed to authorize request", log.ErrorField(err))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var orderID int
		err = render.DecodeJSON(r.Body, &orderID)
		if err != nil {
			logWith.Error("Failed to decode request body", log.ErrorField(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if luna.Valid(orderID) == false {
			logWith.Error("Invalid order number", log.AnyField("order_id", orderID))
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		err = adder.Add(ctx, orderID, user.UUID)
		if err != nil {
			if errors.Is(err, entity.ErrOrderAlreadyUploadedByAnotherUser) {
				w.WriteHeader(http.StatusConflict)
				return
			}
			if errors.Is(err, entity.ErrOrderAlreadyUploaded) {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		logWith.Info(
			"Order successfully added",
			log.AnyField("order_id", orderID),
			log.AnyField("user_uuid", user.UUID),
		)
		w.WriteHeader(http.StatusAccepted)
	}
}

// AllOrdersGetter is an interface for getting an orders from the user.
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=AllOrdersGetter
type AllOrdersGetter interface {
	GetAll(ctx context.Context, userUUID uuid.UUID) ([]entity.Order, error)
}

//TODO swag documentation

// NewAllGetter  returned func for getting all orders from the user.
func NewAllGetter(log *logger.Logger, getter AllOrdersGetter, authorizer UserAuthorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "app.http-server.handler.api.user.orders.NewAllGetter"

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		reqID := middleware.GetReqID(ctx)
		ctx = context.WithValue(ctx, contexter.RequestID, reqID)
		logWith := log.With(
			log.StringField("op", op),
			log.StringField("request_id", reqID),
		)

		user, err := authorizer.Authorize(ctx, r.Header.Get("Authorization"))
		if err != nil {
			logWith.Error("Failed to authorize request", log.ErrorField(err))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		orders, err := getter.GetAll(ctx, user.UUID)
		if err != nil {
			if errors.Is(err, entity.ErrOrderNotFound) {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, orders)
		logWith.Info("Orders successfully retrieved")
		w.WriteHeader(http.StatusOK)
	}
}
