package segment

import (
	"context"
	"errors"
	"net/http"
	"time"

	routeerrs "github.com/adsrkey/dynamic-user-segmentation-service/internal/delivery/http/v1/routes/errors"
	"github.com/adsrkey/dynamic-user-segmentation-service/internal/delivery/http/v1/routes/segment/dto"
	usecase_errors "github.com/adsrkey/dynamic-user-segmentation-service/internal/usecase/errors"
	uc "github.com/adsrkey/dynamic-user-segmentation-service/internal/usecase/segment"
	"github.com/adsrkey/dynamic-user-segmentation-service/pkg/logger"
	"github.com/labstack/echo/v4"
)

type route struct {
	uc  *uc.UseCase
	log logger.Logger
}

func New(g *echo.Group, uc *uc.UseCase, log logger.Logger) {
	route := &route{
		uc:  uc,
		log: log,
	}

	g.POST("", route.create)
	g.DELETE("", route.delete)
}

// create. -
func (r *route) create(c echo.Context) (err error) {
	var (
		// context
		timeout     = 1 * time.Second
		ctx, cancel = context.WithTimeout(c.Request().Context(), timeout)

		// request body dto
		input dto.SlugInput
	)

	defer cancel()

	err = BindSlugInput(c, &input)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrResponse{
			Message: routeerrs.ErrNotDecodeJSONData.Error(),
		})
	}
	err = ValidateSlugInput(c, &input)
	if err != nil {
		return err
	}

	_, err = r.uc.Create(ctx, input.Slug)
	if err != nil {
		if errors.Is(err, usecase_errors.ErrDB) {
			return c.JSON(http.StatusInternalServerError, dto.ErrResponse{
				Message: "InternalServerError",
			})
		}
		return c.JSON(http.StatusConflict, dto.ErrResponse{
			Message: err.Error(),
		})
	}

	return Success(c)
}

func (r *route) delete(c echo.Context) (err error) {
	var (
		// context
		timeout     = 1 * time.Second
		ctx, cancel = context.WithTimeout(c.Request().Context(), timeout)

		// request body dto
		input dto.SlugInput
	)

	defer cancel()

	err = BindSlugInput(c, &input)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrResponse{
			Message: routeerrs.ErrNotDecodeJSONData.Error(),
		})
	}
	err = ValidateSlugInput(c, &input)
	if err != nil {
		return err
	}

	err = r.uc.Delete(ctx, input.Slug)
	if err != nil {
		if errors.Is(err, usecase_errors.ErrDB) {
			return c.JSON(http.StatusInternalServerError, dto.ErrResponse{
				Message: "InternalServerError",
			})
		}
		return c.JSON(http.StatusNotFound, dto.ErrResponse{
			Message: err.Error(),
		})
	}

	return Success(c)
}

func Success(c echo.Context) error {
	return c.JSON(http.StatusCreated, dto.Response{
		Message: "success",
	})
}
