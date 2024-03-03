package withdrawals

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

type BalanceWithdrawOperationGetter interface {
	GetWithdrawOperations(ctx context.Context, userUUID uuid.UUID) ([]entity.BalanceOperation, error)
}

type UserAuthorizer interface {
	Authorize(ctx context.Context, token string) (*entity.User, error)
}

type Request struct {
	OrderNumber string  `json:"order" validate:"required" example:"12312455"`
	Sum         float64 `json:"sum" validate:"required" example:"100"`
}

func New(log *logger.Logger, getter BalanceWithdrawOperationGetter, authorizer UserAuthorizer) http.HandlerFunc {
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
		balance, err := getter.GetWithdrawOperations(ctx, user.UUID)
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
