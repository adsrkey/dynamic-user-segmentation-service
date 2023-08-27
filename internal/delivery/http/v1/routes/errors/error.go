package routeerrs

import "errors"

var (
	ErrNotDecodeJSONData = errors.New("could not decode json data")
)
