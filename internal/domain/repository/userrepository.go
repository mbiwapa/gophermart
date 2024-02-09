package repository

import (
	"context"
	"errors"

	"github.com/mbiwapa/gophermart.git/internal/domain/entity"
)

var (
	// ErrUserNotFound is returned when a user is not found.
	ErrUserNotFound = errors.New("user not found")
	// ErrUserExists is returned when a user already exists.
	ErrUserExists = errors.New("user already exists")
)

// UserRepository is an interface for user repository.
type UserRepository interface {
	// GetUserByUUID(ctx context.Context, id string) (*entity.User, error)
	GetUserByLogin(ctx context.Context, login string) (*entity.User, error)
	CreateUser(ctx context.Context, user *entity.User) (*entity.User, error)
}
