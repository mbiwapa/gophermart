package services

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"

	"github.com/mbiwapa/gophermart.git/internal/domain/entity"
	"github.com/mbiwapa/gophermart.git/internal/domain/repository"
	"github.com/mbiwapa/gophermart.git/internal/domain/tool"
	"github.com/mbiwapa/gophermart.git/internal/lib/logger"
)

// UserService is a service for managing users.
type UserService struct {
	repository repository.UserRepository
	secretKey  string
	logger     *logger.Logger
}

// NewUserService returns a new user service.
func NewUserService(repository repository.UserRepository, logger *logger.Logger, secretKey string) *UserService {
	return &UserService{
		repository: repository,
		secretKey:  secretKey,
		logger:     logger,
	}
}

// Registration register a new user.
func (s *UserService) Registration(ctx context.Context, login, password string) (string, error) {
	const op = "domain.services.UserService.Registration"
	log := s.logger.With(s.logger.StringField("op", op),
		s.logger.StringField("request_id", ctx.Value("request_id").(string)),
	)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Did not generate password hash", log.ErrorField(err))
		return "", err
	}

	user := entity.NewUser(login, string(passwordHash), "", uuid.Nil)

	user, err = s.repository.CreateUser(ctx, user)
	if err != nil {
		return "", err
	}

	JWTstring, err := tool.CreateJWT(user.UUID, s.secretKey)
	if err != nil {
		log.Error("Failed to create JWT", log.ErrorField(err))
		return "", err
	}

	return JWTstring, nil
}

// Authorize authorize a user.
func (s *UserService) Authorize(ctx context.Context, login, password string) (*entity.User, error) {
	const op = "domain.services.UserService.Authorize"
	log := s.logger.With(s.logger.StringField("op", op),
		s.logger.StringField("request_id", ctx.Value("request_id").(string)),
	)

	user, err := s.repository.GetUserByLogin(ctx, login)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		log.Error("Wrong password", log.ErrorField(err))
		return nil, err
	}

	JWTstring, err := tool.CreateJWT(user.UUID, s.secretKey)
	if err != nil {
		log.Error("Failed to create JWT", log.ErrorField(err))
		return nil, err
	}
	user.JWT = JWTstring

	return user, nil
}

// Authenticate authenticates a user.
func (s *UserService) Authenticate(ctx context.Context, token string) (*entity.User, error) {
	const op = "domain.services.UserService.Authenticate"
	log := s.logger.With(s.logger.StringField("op", op),
		s.logger.StringField("request_id", ctx.Value("request_id").(string)),
	)

	userUUID, err := tool.CheckJWT(token, s.secretKey)
	if err != nil || userUUID == uuid.Nil {
		log.Error("Invalid JWT", log.ErrorField(err))
		return nil, err
	}

	user := entity.NewUser("", "", token, userUUID)
	return user, nil
}
