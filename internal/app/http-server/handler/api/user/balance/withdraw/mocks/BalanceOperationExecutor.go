// Code generated by mockery v2.28.2. DO NOT EDIT.

package mocks

import (
	context "context"

	entity "github.com/mbiwapa/gophermart.git/internal/domain/entity"
	mock "github.com/stretchr/testify/mock"
)

// BalanceOperationExecutor is an autogenerated mock type for the BalanceOperationExecutor type
type BalanceOperationExecutor struct {
	mock.Mock
}

// Execute provides a mock function with given fields: ctx, operation
func (_m *BalanceOperationExecutor) Execute(ctx context.Context, operation entity.BalanceOperation) (*entity.Balance, error) {
	ret := _m.Called(ctx, operation)

	var r0 *entity.Balance
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, entity.BalanceOperation) (*entity.Balance, error)); ok {
		return rf(ctx, operation)
	}
	if rf, ok := ret.Get(0).(func(context.Context, entity.BalanceOperation) *entity.Balance); ok {
		r0 = rf(ctx, operation)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*entity.Balance)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, entity.BalanceOperation) error); ok {
		r1 = rf(ctx, operation)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewBalanceOperationExecutor interface {
	mock.TestingT
	Cleanup(func())
}

// NewBalanceOperationExecutor creates a new instance of BalanceOperationExecutor. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewBalanceOperationExecutor(t mockConstructorTestingTNewBalanceOperationExecutor) *BalanceOperationExecutor {
	mock := &BalanceOperationExecutor{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
