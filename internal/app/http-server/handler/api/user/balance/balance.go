package balance

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"github.com/mbiwapa/gophermart.git/internal/domain/entity"
	"github.com/mbiwapa/gophermart.git/internal/lib/contexter"
	"github.com/mbiwapa/gophermart.git/internal/lib/logger"
)

// UserBalanceGetter is an interface for getting an balance from the user.
type UserBalanceGetter interface {
	GetBalance(ctx context.Context, userUUID uuid.UUID) (*entity.Balance, error)
}

// Response is a response for the user balance.
type Response struct {
	Current   float64 `json:"current" example:"500.5"`
	Withdrawn float64 `json:"withdrawn" example:"42"`
}

type UserAuthorizer interface {
	Authorize(ctx context.Context, token string) (*entity.User, error)
}

func New(log *logger.Logger, getter UserBalanceGetter, authorizer UserAuthorizer) http.HandlerFunc {
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
		balance, err := getter.GetBalance(ctx, user.UUID)
		if err != nil {
			_ = balance
			//FIXME add error handling and etc
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		logWith.Info("Balance fetched", log.AnyField("balance", balance))
		w.WriteHeader(http.StatusOK)
	}
}
