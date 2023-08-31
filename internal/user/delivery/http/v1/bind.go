package user

import (
	"net/http"

	dto "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/user"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func BindAddToSegmentInput(c echo.Context, input *dto.AddToSegmentInput) error {
	if err := c.Bind(input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.Wrap(err, "could not decode user data"))
	}
	return nil
}

func BindGetActiveSegmentsInput(c echo.Context, input *dto.GetActiveSegments) error {
	if err := c.Bind(input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.Wrap(err, "could not decode user data"))
	}
	return nil
}

func BindReportInput(c echo.Context, input *dto.ReportInput) error {
	if err := c.Bind(input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.Wrap(err, "could not decode user data"))
	}
	return nil
}
