package balance

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"

	"github.com/mbiwapa/gophermart.git/internal/domain/entity"
	"github.com/mbiwapa/gophermart.git/internal/lib/contexter"
	"github.com/mbiwapa/gophermart.git/internal/lib/logger"
)

// UserBalanceGetter is an interface for getting an balance from the user.
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=UserBalanceGetter
type UserBalanceGetter interface {
	GetBalance(ctx context.Context, userUUID uuid.UUID) (*entity.Balance, error)
}

// UserAuthorizer is an interface for authorizing users.
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=UserAuthorizer
type UserAuthorizer interface {
	Authorize(ctx context.Context, token string) (*entity.User, error)
}

//TODO swag documentation

// New  returned func for getting an balance from the user.
func New(log *logger.Logger, getter UserBalanceGetter, authorizer UserAuthorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "app.http-server.handler.api.user.balance.New"

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

		balance, err := getter.GetBalance(ctx, user.UUID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, balance)
		logWith.Info("Balance fetched")
		w.WriteHeader(http.StatusOK)
	}
}
