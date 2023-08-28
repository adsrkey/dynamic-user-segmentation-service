package segment

import (
	"fmt"
	"net/http"

	dto "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/segment"
	"github.com/labstack/echo/v4"
)

func ValidateSlugInput(c echo.Context, input *dto.SlugInput) error {
	if err := c.Validate(input); err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err)
	}
	if input.Slug == "" {
		return fmt.Errorf("field 'slug' dont be empty")
	}
	return nil
}
