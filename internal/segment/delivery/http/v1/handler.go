package segment

import (
	"context"
	"errors"
	"net/http"
	"time"

	response "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler"
	handler_errors "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/errors"
	dto "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/segment"
	usecase_errors "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/usecase/errors"
	"github.com/adsrkey/dynamic-user-segmentation-service/internal/usecases"
	"github.com/adsrkey/dynamic-user-segmentation-service/pkg/logger"
	"github.com/labstack/echo/v4"
)

type handler struct {
	uc    usecases.Segment
	log   logger.Logger
	group *echo.Group
}

func New(
	group *echo.Group,
	uc usecases.Segment,
	log logger.Logger) *handler { // TODO: to interface

	h := &handler{
		uc:    uc,
		log:   log,
		group: group,
	}
	h.MapUserRoutes()

	return h
}

// create. -
func (r *handler) create(c echo.Context) (err error) {
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
		return c.JSON(http.StatusBadRequest, response.ErrResponse{
			Message: handler_errors.ErrNotDecodeJSONData.Error(),
		})
	}
	err = ValidateSlugInput(c, &input)
	if err != nil {
		return err
	}

	_, err = r.uc.Create(ctx, input.Slug)
	if err != nil {
		if errors.Is(err, usecase_errors.ErrDB) {
			return c.JSON(http.StatusInternalServerError, response.ErrResponse{
				Message: "InternalServerError",
			})
		}
		return c.JSON(http.StatusConflict, response.ErrResponse{
			Message: err.Error(),
		})
	}

	return Success(c)
}

func (r *handler) delete(c echo.Context) (err error) {
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
		return c.JSON(http.StatusBadRequest, response.ErrResponse{
			Message: handler_errors.ErrNotDecodeJSONData.Error(),
		})
	}
	err = ValidateSlugInput(c, &input)
	if err != nil {
		return err
	}

	err = r.uc.Delete(ctx, input.Slug)
	if err != nil {
		if errors.Is(err, usecase_errors.ErrDB) {
			return c.JSON(http.StatusInternalServerError, response.ErrResponse{
				Message: "InternalServerError",
			})
		}
		return c.JSON(http.StatusNotFound, response.ErrResponse{
			Message: err.Error(),
		})
	}

	return Success(c)
}

func Success(c echo.Context) error {
	return c.JSON(http.StatusCreated, response.Response{
		Message: "success",
	})
}
