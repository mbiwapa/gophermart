package login

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"

	"github.com/mbiwapa/gophermart.git/internal/domain/user/entity"
	"github.com/mbiwapa/gophermart.git/internal/lib/contexter"
	"github.com/mbiwapa/gophermart.git/internal/lib/logger"
)

// UserAuthenticator is an interface for user authentication.
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=UserAuthenticator
type UserAuthenticator interface {
	Authenticate(ctx context.Context, login, password string) (string, error)
}

// Request struct for HTTP Request in JSON
type Request struct {
	Login    string `json:"login" validate:"required" example:"test@test.com"`
	Password string `json:"password" validate:"required" example:"test_Password"`
}

// New returned func for logging in a user.
//
//	@Summary		Аутентификация пользователя.
//	@Description	Эндпоинт используется для аутентификации пользователя.
//	@Description	Логин приводится к нижнему регистру на стороне сервера
//	@Description	В заголовке Authorization возвращается JWT токен для авторизации
//	@Accept			json
//	@Router			/user/login [post]
//	@Param			login		body	login.Request	true	"Login of the user"
//	@Param			password	body	login.Request	true	"Password of the user"
//	@Success		200			"User successfully authenticated"
//	@Failure		401			"Login or password is wrong"
//	@Failure		400			"Bad request"
//	@Failure		500			"Internal server error"
//	@Header			200			{string}	Authorization	"token"
func New(log *logger.Logger, service UserAuthenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "app.http-server.handler.api.user.login.New"

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		reqID := middleware.GetReqID(ctx)
		ctx = context.WithValue(ctx, contexter.RequestId, reqID)
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
		//convert Login to lower case
		req.Login = strings.ToLower(req.Login)
		log.Info("Request decoded", log.AnyField("login", req.Login))

		if err := validator.New().Struct(req); err != nil {
			log.Error("Failed to validate request body", log.ErrorField(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		jwtString, err := service.Authenticate(ctx, req.Login, req.Password)
		if err != nil {
			if errors.Is(err, entity.ErrUserWrongPassword) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Authorization", jwtString)
		log.Info("User Authenticated")
		w.WriteHeader(http.StatusOK)
	}
}
