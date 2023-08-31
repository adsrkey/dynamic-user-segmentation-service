package segment

import (
	"net/http"

	dto "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/segment"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func BindSegmentAddInput(c echo.Context, input *dto.SegmentAddInput) error {
	if err := c.Bind(input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.Wrap(err, "could not decode user data"))
	}
	return nil
}

func BindSegmentDelInput(c echo.Context, input *dto.SegmentDelInput) error {
	if err := c.Bind(input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.Wrap(err, "could not decode user data"))
	}
	return nil
}
