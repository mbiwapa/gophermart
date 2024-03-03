package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"github.com/mbiwapa/gophermart.git/internal/domain/entity"
	"github.com/mbiwapa/gophermart.git/internal/domain/service"
	"github.com/mbiwapa/gophermart.git/internal/domain/service/mocks"
	"github.com/mbiwapa/gophermart.git/internal/domain/tool"
	"github.com/mbiwapa/gophermart.git/internal/lib/contexter"
	"github.com/mbiwapa/gophermart.git/internal/lib/logger"
)

func TestUserService_Registration(t *testing.T) {
	ctx := context.WithValue(context.Background(), contexter.RequestID, "req")
	type args struct {
		ctx      context.Context
		login    string
		password string
	}
	tests := []struct {
		name string

		args           args
		want           string
		wantErr        error
		repositoryWant *entity.User
		repositoryErr  error
	}{
		{
			name: "Create a new user: success",
			args: args{
				ctx:      ctx,
				login:    "test",
				password: "password",
			},
			want:           "jwtString",
			wantErr:        nil,
			repositoryWant: entity.NewUser("", "", "", uuid.New()),
			repositoryErr:  nil,
		},
		{
			name: "Create a new user: User already exists",
			args: args{
				ctx:      ctx,
				login:    "test",
				password: "password",
			},
			want:           "",
			wantErr:        entity.ErrUserExists,
			repositoryWant: nil,
			repositoryErr:  entity.ErrUserExists,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			//Prepare mocks
			NewUserRepositoryMock := mocks.NewUserRepository(t)
			NewUserRepositoryMock.On("CreateUser", mock.Anything, mock.Anything).
				Return(tc.repositoryWant, tc.repositoryErr).
				Once()
			log := logger.NewLogger()
			s := service.NewUserService(NewUserRepositoryMock, log, "secret")
			//Body of test
			got, err := s.Register(tc.args.ctx, tc.args.login, tc.args.password)
			//Asserts
			assert.Equal(t, tc.wantErr, err)
			if tc.wantErr == nil {
				assert.IsType(t, tc.want, got)
			} else {
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestUserService_Authenticate(t *testing.T) {
	ctx := context.WithValue(context.Background(), contexter.RequestID, "req")
	password := "password"
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Did not generate password hash: %v", err)
	}
	type args struct {
		ctx      context.Context
		login    string
		password string
	}
	tests := []struct {
		name string

		args           args
		want           string
		wantErr        error
		repositoryWant *entity.User
		repositoryErr  error
	}{
		{
			name: "Authenticate: Success",
			args: args{
				ctx:      ctx,
				login:    "test",
				password: "password",
			},
			want:           "jwtString",
			wantErr:        nil,
			repositoryWant: &entity.User{UUID: uuid.New(), Login: "test", PasswordHash: string(passwordHash), JWT: "jwtString"},
			repositoryErr:  nil,
		},
		{
			name: "Authenticate a user: wrong password",
			args: args{
				ctx:      ctx,
				login:    "test",
				password: "wrong",
			},
			want:           "",
			wantErr:        entity.ErrUserWrongPassword,
			repositoryWant: &entity.User{UUID: uuid.New(), Login: "test", PasswordHash: string(passwordHash), JWT: "jwtString"},
			repositoryErr:  nil,
		},
		{
			name: "Authenticate a user: user not found",
			args: args{
				ctx:      ctx,
				login:    "wrong",
				password: "password",
			},
			want:           "",
			wantErr:        entity.ErrUserNotFound,
			repositoryWant: nil,
			repositoryErr:  entity.ErrUserNotFound,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			//Prepare mocks
			NewUserRepositoryMock := mocks.NewUserRepository(t)
			NewUserRepositoryMock.On("GetUserByLogin", mock.Anything, mock.Anything).
				Return(tc.repositoryWant, tc.repositoryErr).
				Once()
			log := logger.NewLogger()
			s := service.NewUserService(NewUserRepositoryMock, log, "secret")
			//Body of test
			got, err := s.Authenticate(tc.args.ctx, tc.args.login, tc.args.password)
			//Asserts
			assert.Equal(t, tc.wantErr, err)
			if tc.wantErr == nil {
				assert.IsType(t, tc.want, got)
			} else {
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestUserService_Authorize(t *testing.T) {
	ctx := context.WithValue(context.Background(), contexter.RequestID, "req")
	userUUID, err := uuid.NewRandom()
	if err != nil {
		t.Fatalf("Failed to create user UUID: %v", err)
	}
	jwtString, err := tool.CreateJWT(userUUID, "secret")

	if err != nil {
		t.Fatalf("Failed to create JWT: %v", err)
	}
	type args struct {
		ctx   context.Context
		token string
	}
	tests := []struct {
		name    string
		args    args
		want    *entity.User
		wantErr bool
	}{
		{
			name: "Authorize a user: success",
			args: args{
				ctx:   ctx,
				token: jwtString,
			},
			want: &entity.User{
				UUID:  userUUID,
				Login: "",
				JWT:   jwtString,
			},
			wantErr: false,
		},
		{
			name: "Authorize a user: invalid JWT",
			args: args{
				ctx:   ctx,
				token: "invalid",
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			//Prepare mocks
			NewUserRepositoryMock := mocks.NewUserRepository(t)
			log := logger.NewLogger()
			s := service.NewUserService(NewUserRepositoryMock, log, "secret")
			//Body of test
			got, err := s.Authorize(tc.args.ctx, tc.args.token)
			//Asserts
			assert.Equal(t, tc.want, got)
			if tc.wantErr {
				assert.Error(t, err)
			}
		})
	}
}
