package server

import (
	"log"
	"net/http"
	"os"

	segmentHandler "github.com/adsrkey/dynamic-user-segmentation-service/internal/segment/delivery/http/v1"
	"github.com/adsrkey/dynamic-user-segmentation-service/internal/usecases"
	userHandler "github.com/adsrkey/dynamic-user-segmentation-service/internal/user/delivery/http/v1"
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

func (de *Delivery) MapRoutes(usecases usecases.UseCases) {
	// TODO: Add Middleware
	de.echo.GET("/health", func(c echo.Context) error { return c.NoContent(200) })

	v1 := de.echo.Group("/api/v1")
	{
		v1.Use(ValidContentType)
		segmentGroup := v1.Group("/segment")
		{
			segmentHandler.New(segmentGroup, usecases.Segment(), de.echo.Logger)
		}
		userGroup := v1.Group("/user")
		{
			userHandler.New(userGroup, usecases.User(), de.echo.Logger)
		}
	}
}

func setLogsFile() *os.File {
	file, err := os.OpenFile("/logs/requests.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal(err)
	}
	return file
}
