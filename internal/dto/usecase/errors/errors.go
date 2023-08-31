package usecase_errors

import "errors"

var (
	ErrDB      = errors.New("db error")
	ToFewUsers = errors.New("too few users to automatically add a segment")
)
