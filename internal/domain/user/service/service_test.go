package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mbiwapa/gophermart.git/internal/domain/user/entity"
	"github.com/mbiwapa/gophermart.git/internal/domain/user/service"
	"github.com/mbiwapa/gophermart.git/internal/domain/user/service/mocks"
	"github.com/mbiwapa/gophermart.git/internal/lib/logger"
)

func TestUserService_Registration(t *testing.T) {

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
				ctx:      context.WithValue(context.Background(), "request_id", "req"),
				login:    "test",
				password: "password",
			},
			want:           "jwtstring",
			wantErr:        nil,
			repositoryWant: entity.NewUser("", "", "", uuid.New()),
			repositoryErr:  nil,
		},
		{
			name: "Create a new user: User already exists",
			args: args{
				ctx:      context.WithValue(context.Background(), "request_id", "req"),
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
			got, err := s.Registration(tc.args.ctx, tc.args.login, tc.args.password)
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
