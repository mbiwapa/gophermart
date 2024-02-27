package entity

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	New        Status = "NEW"
	Processing Status = "PROCESSING"
	Invalid    Status = "INVALID"
	Processed  Status = "PROCESSED"
	Registered Status = "REGISTERED" //Only for CLIENT
)

// Order is an entity for managing orders.
type Order struct {
	Number  string  `json:"number" example:"123124551"`
	Status  Status  `json:"status" example:"PROCESSING"`
	Accrual float64 `json:"accrual,omitempty" example:"500"`
	// ignore in json
	UserUUID uuid.UUID `json:"-"`
	//time format RFC3339
	UploadedAt time.Time `json:"uploaded_at" example:"2020-12-10T15:15:45+03:00"`
}

//FIXME добавить ошибки и фабрику

func NewOrder(userUUID uuid.UUID, orderNumber string) Order {

	order := Order{
		UserUUID:   userUUID,
		Number:     orderNumber,
		UploadedAt: time.Now(),
		Status:     New,
	}

	return order
}
