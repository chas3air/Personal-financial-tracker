package models

import "github.com/google/uuid"

type User struct {
	Id       uuid.UUID `validate:"required"`
	Login    string    `validate:"required"`
	Password string    `validate:"required"`
	Role     string    `validate:"required"`
}
