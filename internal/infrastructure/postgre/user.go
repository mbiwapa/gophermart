package postgre

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/mbiwapa/gophermart.git/internal/domain/entity"
	"github.com/mbiwapa/gophermart.git/internal/domain/repository"
	"github.com/mbiwapa/gophermart.git/internal/lib/logger"
)

// UserRepository is an implementation of user repository.
type UserRepository struct {
	db  *sql.DB
	log *logger.Logger
}

// NewUserRepository returns a new postgre user repository
func NewUserRepository(db *sql.DB, log *logger.Logger) (*UserRepository, error) {
	const op = "infrastructure.postgre.NewUserRepository"

	logger := log.With(log.StringField("op", op))

	storage := &UserRepository{db: db, log: log}

	stmt, err := db.Prepare(`CREATE TABLE IF NOT EXISTS users (
		uuid UUID PRIMARY KEY,
        login TEXT UNIQUE NOT NULL,
        password_hash TEXT NOT NULL);`)
	if err != nil {
		logger.Error("Failed to create table", log.ErrorField(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return storage, nil
}

// GetUserByUUID returns a user by ID.
// func (r *UserRepository) GetUserByUUID(ctx context.Context, id string) (*entity.User, error) {
// 	const op = "infrastructure.postgre.UserRepository.GetUserByID"
// 	logger := r.log.With(r.log.StringField("op", op), r.log.StringField("user_id", id))
// 	logger.Info("Getting user by id")

// 	var user entity.User

// 	stmt, err := r.db.PrepareContext(ctx, "SELECT uuid, login, password_hash FROM users WHERE uuid = $1")
// 	if err != nil {
// 		logger.Error("Failed to prepare statement", logger.ErrorField(err))
// 		return nil, fmt.Errorf("%s: %w", op, err)
// 	}
// 	defer stmt.Close()

// 	err = stmt.QueryRowContext(ctx, id).Scan(&user.UUID, &user.Login, &user.PasswordHash)
// 	if err != nil {
// 		if errors.Is(err, pgx.ErrNoRows) {
// 			logger.Info("User not found", logger.ErrorField(err))
// 			return nil, repository.ErrUserNotFound
// 		}
// 		logger.Error("Failed to get user", logger.ErrorField(err))
// 		return nil, fmt.Errorf("%s: %w", op, err)
// 	}
// 	return &user, nil
// }

// GetUserByLogin returns a user by login.
func (r *UserRepository) GetUserByLogin(ctx context.Context, login string) (*entity.User, error) {
	const op = "infrastructure.postgre.UserRepository.GetUserByLogin"
	logger := r.log.With(r.log.StringField("op", op), r.log.StringField("user_login", login))

	var user entity.User
	stmt, err := r.db.PrepareContext(ctx, "SELECT uuid, login, password_hash FROM users WHERE login = $1")
	if err != nil {
		logger.Error("Failed to prepare statement", logger.ErrorField(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()
	//get user by login
	err = stmt.QueryRowContext(ctx, login).Scan(&user.UUID, &user.Login, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Info("User not found", logger.ErrorField(err))
			return nil, repository.ErrUserNotFound
		}
		logger.Error("Failed to get user", logger.ErrorField(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &user, nil

}

// CreateUser creates a new user.
func (r *UserRepository) CreateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	const op = "infrastructure.postgre.UserRepository.CreateUser"

	logger := r.log.With(
		r.log.StringField("op", op),
		r.log.StringField("request_id", ctx.Value("request_id").(string)),
		r.log.StringField("user_login", user.Login),
	)

	stmt, err := r.db.PrepareContext(ctx, "INSERT INTO users (uuid, login, password_hash) VALUES ($1, $2, $3)")
	if err != nil {
		logger.Error("Failed to prepare statement", logger.ErrorField(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, user.UUID, user.Login, user.PasswordHash)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				logger.Info("User already exists!")
				return nil, repository.ErrUserExists
			}
		}
		logger.Error("Failed to create user", logger.ErrorField(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return user, nil
}
