// Code generated by mockery v2.28.2. DO NOT EDIT.

package mocks

import (
	context "context"

	entity "github.com/mbiwapa/gophermart.git/internal/domain/entity"
	mock "github.com/stretchr/testify/mock"

	uuid "github.com/google/uuid"
)

// UserBalanceGetter is an autogenerated mock type for the UserBalanceGetter type
type UserBalanceGetter struct {
	mock.Mock
}

// GetBalance provides a mock function with given fields: ctx, userUUID
func (_m *UserBalanceGetter) GetBalance(ctx context.Context, userUUID uuid.UUID) (*entity.Balance, error) {
	ret := _m.Called(ctx, userUUID)

	var r0 *entity.Balance
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*entity.Balance, error)); ok {
		return rf(ctx, userUUID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *entity.Balance); ok {
		r0 = rf(ctx, userUUID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*entity.Balance)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, userUUID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewUserBalanceGetter interface {
	mock.TestingT
	Cleanup(func())
}

// NewUserBalanceGetter creates a new instance of UserBalanceGetter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewUserBalanceGetter(t mockConstructorTestingTNewUserBalanceGetter) *UserBalanceGetter {
	mock := &UserBalanceGetter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
