package v1

import (
	"log"
	"net/http"
	"os"

	segmentRoute "github.com/adsrkey/dynamic-user-segmentation-service/internal/delivery/http/v1/routes/segment"
	userRoute "github.com/adsrkey/dynamic-user-segmentation-service/internal/delivery/http/v1/routes/user"
	"github.com/adsrkey/dynamic-user-segmentation-service/internal/usecase"

	"github.com/labstack/echo/v4"
)

func ExecTime(next echo.HandlerFunc) echo.HandlerFunc {
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

func New(e *echo.Echo, usecase usecase.UseCases) {
	// TODO: Add Middleware
	e.GET("/health", func(c echo.Context) error { return c.NoContent(200) })

	v1 := e.Group("/api/v1")
	{
		// v1.Use(ExecTime)
		segmentGroup := v1.Group("/segment")
		{
			segmentRoute.New(segmentGroup, usecase.Segment(), e.Logger)
		}
		userGroup := v1.Group("/user")
		{
			userRoute.New(userGroup, usecase.User(), e.Logger)
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
