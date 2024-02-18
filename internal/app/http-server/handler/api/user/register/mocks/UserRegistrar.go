// Code generated by mockery v2.28.2. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// UserRegistrar is an autogenerated mock type for the UserRegistrar type
type UserRegistrar struct {
	mock.Mock
}

// Register provides a mock function with given fields: ctx, login, password
func (_m *UserRegistrar) Register(ctx context.Context, login string, password string) (string, error) {
	ret := _m.Called(ctx, login, password)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (string, error)); ok {
		return rf(ctx, login, password)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) string); ok {
		r0 = rf(ctx, login, password)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, login, password)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewUserRegistrar interface {
	mock.TestingT
	Cleanup(func())
}

// NewUserRegistrar creates a new instance of UserRegistrar. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewUserRegistrar(t mockConstructorTestingTNewUserRegistrar) *UserRegistrar {
	mock := &UserRegistrar{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}