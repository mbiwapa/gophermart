package service

import (
	"context"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/mbiwapa/gophermart.git/internal/domain/entity"
	"github.com/mbiwapa/gophermart.git/internal/domain/tool"
	"github.com/mbiwapa/gophermart.git/internal/lib/contexter"
	"github.com/mbiwapa/gophermart.git/internal/lib/logger"
)

// UserRepository is an interface for user repository.
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=UserRepository
type UserRepository interface {
	GetUserByLogin(ctx context.Context, login string) (*entity.User, error)
	CreateUser(ctx context.Context, user *entity.User) (*entity.User, error)
	GetUserByUUID(ctx context.Context, userUUID uuid.UUID) (*entity.User, error)
}

// UserService is a service for managing users.
type UserService struct {
	repository UserRepository
	secretKey  string
	logger     *logger.Logger
}

// NewUserService returns a new user service.
func NewUserService(repository UserRepository, logger *logger.Logger, secretKey string) *UserService {
	return &UserService{
		repository: repository,
		secretKey:  secretKey,
		logger:     logger,
	}
}

// Register register a new user.
func (s *UserService) Register(ctx context.Context, login, password string) (string, error) {
	const op = "domain.services.UserService.Registration"
	log := s.logger.With(
		s.logger.StringField("op", op),
		s.logger.StringField("request_id", contexter.GetRequestID(ctx)),
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

	jwtString, err := tool.CreateJWT(user.UUID, s.secretKey)
	if err != nil {
		log.Error("Failed to create JWT", log.ErrorField(err))
		return "", err
	}

	return jwtString, nil
}

// Authenticate authorize a user.
func (s *UserService) Authenticate(ctx context.Context, login, password string) (string, error) {
	const op = "domain.services.UserService.Authorize"
	log := s.logger.With(s.logger.StringField("op", op),
		s.logger.StringField("request_id", contexter.GetRequestID(ctx)),
	)

	user, err := s.repository.GetUserByLogin(ctx, login)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		log.Error("Wrong password", log.ErrorField(err))
		return "", entity.ErrUserWrongPassword
	}

	jwtString, err := tool.CreateJWT(user.UUID, s.secretKey)
	if err != nil {
		log.Error("Failed to create JWT", log.ErrorField(err))
		return "", err
	}

	return jwtString, nil
}

// Authorize authenticates a user.
func (s *UserService) Authorize(ctx context.Context, token string) (*entity.User, error) {
	const op = "domain.services.UserService.Authenticate"
	log := s.logger.With(s.logger.StringField("op", op),
		s.logger.StringField("request_id", contexter.GetRequestID(ctx)),
	)

	userUUID, err := tool.CheckJWT(token, s.secretKey)
	if err != nil || userUUID == uuid.Nil {
		log.Error("Invalid JWT", log.ErrorField(err))
		return nil, err
	}

	user := entity.NewUser("", "", token, userUUID)
	return user, nil
}
