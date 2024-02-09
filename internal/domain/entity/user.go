package entity

import (
	"github.com/google/uuid"
)

// User is an aggregate for managing users.
type User struct {
	UUID         uuid.UUID
	Login        string
	PasswordHash string
	JWT          string
}

// NewUser returns a new user.
func NewUser(login, passwordHash, jwtToken string, userUUID uuid.UUID) *User {

	user := &User{}

	if userUUID != uuid.Nil {
		user.UUID = userUUID
	} else {
		user.UUID = uuid.New()
	}
	user.Login = login
	user.PasswordHash = passwordHash
	user.JWT = jwtToken

	return user
}
