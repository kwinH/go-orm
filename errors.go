package oorm

import (
	"errors"
)

var (
	ErrNotFind          = errors.New("not find")
	ErrParam            = errors.New("parameter error")
	ErrMissingCondition = errors.New("missing condition")
	ErrMissingTableName = errors.New("missing table name")
	ErrInvalidDB        = errors.New("invalid db")
)
