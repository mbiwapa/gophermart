package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Balance struct {
	UserUUID uuid.UUID `json:"-"`
	Current  float64   `json:"current" example:"500.5"`
	Withdraw float64   `json:"withdrawn,omitempty" example:"42"`
}

type BalanceOperation struct {
	UUID        uuid.UUID `json:"-"`
	UserUUID    uuid.UUID `json:"-"`
	Accrual     float64   `json:"-"`
	Withdrawal  float64   `json:"sum" validate:"required" example:"100"`
	OrderNumber int       `json:"order" validate:"required" example:"12312455"`
	ProcessedAt time.Time `json:"processed_at" validate:"required" example:"2020-12-10T15:15:45+03:00"`
}

var (
	// ErrBalanceInsufficientFunds there are insufficient funds in the account
	ErrBalanceInsufficientFunds = errors.New("insufficient funds in the account")
	// ErrBalanceOperationsNotFound there are no balance operations
	ErrBalanceOperationsNotFound = errors.New("balance operations not found")
)

func NewBalanceOperation(userUUID uuid.UUID, accrual, withdrawal float64, orderNumber int) BalanceOperation {
	var operation = BalanceOperation{}

	operation.UUID = uuid.New()
	operation.UserUUID = userUUID
	operation.Accrual = accrual
	operation.Withdrawal = withdrawal
	operation.OrderNumber = orderNumber
	operation.ProcessedAt = time.Now()
	return operation
}
