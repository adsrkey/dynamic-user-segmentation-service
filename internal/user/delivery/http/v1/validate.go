package user

import (
	"fmt"
	"net/http"

	dto "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/user"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func ValidateAddToSegmentInput(c echo.Context, input *dto.AddToSegmentInput) error {
	if err := c.Validate(input); err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err)
	}
	if len(input.SlugsAdd) == 0 && len(input.SlugsDel) == 0 {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, fmt.Errorf("field 'slugs_XXX' array dont be empty"))
	}
	return nil
}

func ValidateGetActiveSegmentsInput(c echo.Context, input *dto.GetActiveSegments) error {
	if err := c.Validate(input); err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err)
	}
	if input.UserID == uuid.Nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, fmt.Errorf("field 'user_id' dont be empty"))
	}
	return nil
}

func ValidateReportInput(c echo.Context, input *dto.ReportInput) error {
	if err := c.Validate(input); err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err)
	}
	if input.UserID == uuid.Nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, fmt.Errorf("fields dont be empty"))
	}
	if input.Year == 0 {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, fmt.Errorf("field 'year' dont be empty"))
	}
	if input.Month == 0 {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, fmt.Errorf("field 'month' dont be empty"))
	}
	return nil
}
