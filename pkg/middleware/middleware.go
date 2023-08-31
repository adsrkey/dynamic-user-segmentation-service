package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func ValidContentType(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		contentType := c.Request().Header.Get(echo.HeaderContentType)
		if contentType != echo.MIMEApplicationJSON {

			type errResponse struct {
				Message string `json:"message"`
			}

			if contentType == "" {
				msg := "missing Header: Content-Type"
				c.JSON(http.StatusBadRequest, errResponse{
					Message: msg,
				})
				return nil
			}

			msg := "invalid mime type: " + contentType

			c.JSON(http.StatusBadRequest, errResponse{
				Message: msg,
			})
			return nil
		}

		if err := next(c); err != nil {
			c.Error(err)
		}
		return nil
	}
}
