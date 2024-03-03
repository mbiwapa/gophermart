package entity

import (
	"time"

	"github.com/google/uuid"
)

type Balance struct {
	UserUUID uuid.UUID
	Current  float64
	Withdraw float64
}

type BalanceOperation struct {
	UUID        uuid.UUID
	UserUUID    uuid.UUID
	Accrual     float64
	Withdrawal  float64
	OrderNumber string
	ProcessedAt time.Time
}
