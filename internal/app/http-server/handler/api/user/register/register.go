package register

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"

	"github.com/mbiwapa/gophermart.git/internal/domain/repository"
	"github.com/mbiwapa/gophermart.git/internal/lib/logger"
)

// UserRegistrator is an interface for user registrator.
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=UserRegistrator
type UserRegistrator interface {
	Registration(ctx context.Context, login, password string) (string, error)
}

// Request struct for HTTP Request in JSON
type Request struct {
	Login    string `json:"login" validate:"required,lowercase"`
	Password string `json:"password" validate:"required"`
}

// New returned func for save new url
func New(log *logger.Logger, service UserRegistrator) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		const op = "app.http-server.handler.api.user.register.New"

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		reqID := middleware.GetReqID(ctx)
		ctx = context.WithValue(ctx, "request_id", reqID)
		log = log.With(
			log.StringField("op", op),
			log.StringField("request_id", reqID),
		)

		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("Failed to decode request body", log.ErrorField(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Info("Request decoded", log.AnyField("login", req.Login))

		if err := validator.New().Struct(req); err != nil {
			log.Error("Failed to validate request body", log.ErrorField(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		jwtString, err := service.Registration(ctx, req.Login, req.Password)
		if err != nil {
			if err == repository.ErrUserExists {
				w.WriteHeader(http.StatusConflict)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Authorization", jwtString)
		log.Info("User registered")
		w.WriteHeader(http.StatusOK)

	}
}
