package handler_errors

import "errors"

var (
	ErrNotDecodeJSONData = errors.New("could not decode json data")
)
