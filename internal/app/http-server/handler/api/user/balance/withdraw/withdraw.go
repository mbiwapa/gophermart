package withdraw

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"

	"github.com/mbiwapa/gophermart.git/internal/domain/entity"
	"github.com/mbiwapa/gophermart.git/internal/lib/contexter"
	"github.com/mbiwapa/gophermart.git/internal/lib/logger"
	"github.com/mbiwapa/gophermart.git/internal/lib/luna"
)

// BalanceOperationExecutor is an interface for withdrawing money from the user's balance.'
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=BalanceOperationExecutor
type BalanceOperationExecutor interface {
	Execute(ctx context.Context, operation entity.BalanceOperation) error
}

// UserAuthorizer is an interface for authorizing users.
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=UserAuthorizer
type UserAuthorizer interface {
	Authorize(ctx context.Context, token string) (*entity.User, error)
}

// Request is a struct for request withdrawing money from the user's balance.
type Request struct {
	OrderNumber string  `json:"order" validate:"required" example:"12312455"`
	Sum         float64 `json:"sum" validate:"required" example:"100"`
}

// New  returned func for withdrawing money from the user's balance.
func New(log *logger.Logger, executor BalanceOperationExecutor, authorizer UserAuthorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "app.http-server.handler.api.user.balance.withdraw.New"

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
			logWith.Info("Failed to authorize request", log.ErrorField(err))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var request Request
		err = render.DecodeJSON(r.Body, &request)
		if err != nil {
			logWith.Info("Failed to decode request body", log.ErrorField(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := validator.New().Struct(request); err != nil {
			logWith.Info("Failed to validate request body", log.ErrorField(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var orderNumber int
		orderNumber, err = strconv.Atoi(request.OrderNumber)
		if err != nil {
			logWith.Info("Invalid order number", log.AnyField("order_id", request.OrderNumber))
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		if luna.Valid(orderNumber) == false {
			logWith.Info("Invalid order number", log.AnyField("order_id", request.OrderNumber))
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		operation := entity.NewBalanceOperation(user.UUID, 0, request.Sum, orderNumber)
		err = executor.Execute(ctx, operation)
		if err != nil {
			if errors.Is(err, entity.ErrBalanceInsufficientFunds) {
				w.WriteHeader(http.StatusPaymentRequired)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		logWith.Info("Successfully executed operation")
		w.WriteHeader(http.StatusOK)
	}
}
