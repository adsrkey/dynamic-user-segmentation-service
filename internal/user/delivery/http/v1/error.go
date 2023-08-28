package user

import (
	"errors"
	"fmt"

	"github.com/labstack/echo/v4"
)

var (
	ErrInvalidAuthHeader = fmt.Errorf("invalid header")
	ErrCannotParseToken  = fmt.Errorf("cannot parse")
)

func newErrorResponse(c echo.Context, errStatus int, message string) {
	err := errors.New(message)
	_, ok := err.(*echo.HTTPError)
	if !ok {
		report := echo.NewHTTPError(errStatus, err.Error())
		_ = c.JSON(errStatus, report)
	}
	c.Error(errors.New("internal server error"))
}
