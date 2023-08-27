package repoerrs

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrDoesNotExist  = errors.New("does not exist")

	ErrDB = errors.New("db error")

	ErrDatabaseConnection = errors.New("db connection error")
)
