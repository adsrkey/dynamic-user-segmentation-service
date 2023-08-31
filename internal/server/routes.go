package server

import (
	"log"
	"os"

	segmentHandler "github.com/adsrkey/dynamic-user-segmentation-service/internal/segment/delivery/http/v1"
	"github.com/adsrkey/dynamic-user-segmentation-service/internal/usecases"
	userHandler "github.com/adsrkey/dynamic-user-segmentation-service/internal/user/delivery/http/v1"
	"github.com/labstack/echo/v4"
)

func (de *Delivery) MapRoutes(usecases usecases.UseCases) {
	de.echo.GET("/health", func(c echo.Context) error { return c.NoContent(200) })

	v1 := de.echo.Group("/api/v1")
	{
		segmentGroup := v1.Group("/segments")
		{
			segmentHandler.New(segmentGroup, usecases.Segment(), de.echo.Logger)
		}
		userGroup := v1.Group("/users")
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
