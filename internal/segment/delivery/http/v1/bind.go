package segment

import (
	"net/http"

	dto "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/segment"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func BindSlugInput(c echo.Context, input *dto.SlugInput) error {
	if err := c.Bind(input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.Wrap(err, "could not decode user data"))
	}
	return nil
}
