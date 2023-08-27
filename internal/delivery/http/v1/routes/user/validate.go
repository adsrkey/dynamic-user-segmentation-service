package user

import (
	"fmt"
	"net/http"

	"github.com/adsrkey/dynamic-user-segmentation-service/internal/delivery/http/v1/routes/user/dto"
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
	return nil
}
