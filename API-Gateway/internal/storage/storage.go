package storageerrors

import (
	"errors"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrAlreadyExists   = errors.New("already exists")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrDeadlineExeeced = errors.New("deadline exceeded")
	ErrContextCanceled = errors.New("context canceled")
	ErrInternal        = errors.New("internal")
)
