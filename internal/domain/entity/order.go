package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	OrderNew        Status = "NEW"
	OrderProcessing Status = "PROCESSING"
	OrderInvalid    Status = "INVALID"
	OrderProcessed  Status = "PROCESSED"
	// OrderRegistered Status use only in external system
	OrderRegistered Status = "REGISTERED"
)

// Order is an entity for managing orders.
type Order struct {
	Number  int     `json:"number" example:"123124551"`
	Status  Status  `json:"status" example:"PROCESSING"`
	Accrual float64 `json:"accrual,omitempty" example:"500"`
	// ignore in json
	UserUUID uuid.UUID `json:"-"`
	//time format RFC3339
	UploadedAt time.Time `json:"uploaded_at" example:"2020-12-10T15:15:45+03:00"`
}

var (
	// ErrOrderAlreadyUploaded is returned when an order is already uploaded.
	ErrOrderAlreadyUploaded = errors.New("order already uploaded")
	// ErrOrderAlreadyUploadedByAnotherUser is returned when an order is already uploaded by another user.
	ErrOrderAlreadyUploadedByAnotherUser = errors.New("order already uploaded by another user")
	// ErrOrderNotFound is returned when an order is not found.
	ErrOrderNotFound = errors.New("order not found")

	// ErrExternalOrderNotRegistered is returned when an order is not registered in external system.
	ErrExternalOrderNotRegistered = errors.New("external order not registered")
	// ErrExternalOrderRateLimitExceeded is returned when an order is rate limit exceeded in external system.
	ErrExternalOrderRateLimitExceeded = errors.New("external order rate limit exceeded")
)

func NewOrder(userUUID uuid.UUID, orderNumber int) Order {

	order := Order{
		UserUUID:   userUUID,
		Number:     orderNumber,
		UploadedAt: time.Now(),
		Status:     OrderNew,
	}

	return order
}
