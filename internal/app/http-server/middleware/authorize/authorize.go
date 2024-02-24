package authorize

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/mbiwapa/gophermart.git/internal/domain/user/entity"
	"github.com/mbiwapa/gophermart.git/internal/lib/contexter"
	"github.com/mbiwapa/gophermart.git/internal/lib/logger"
)

// UserAuthorizer is an interface for user authentication.
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=UserAuthorizer
type UserAuthorizer interface {
	Authorize(ctx context.Context, token string) (*entity.User, error)
}

// New returns a new http.Handler that authorizes requests.
func New(log *logger.Logger, service UserAuthorizer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		const op = "internal.app.http-server.middleware.authorize.New"
		logWith := log.With(
			log.StringField("op", op),
		)
		log.Info("Authorize middleware enabled")

		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
			defer cancel()

			reqID := middleware.GetReqID(ctx)
			ctx = context.WithValue(ctx, contexter.RequestID, reqID)
			logWith = logWith.With(
				log.StringField("request_id", reqID),
			)

			jwtString := r.Header.Get("Authorization")
			if jwtString == "" {
				logWith.Error("Authorization header is not set")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			_, err := service.Authorize(ctx, jwtString)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
