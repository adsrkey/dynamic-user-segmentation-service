package segment

import (
	"context"
	"errors"
	"net/http"
	"time"

	response "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler"
	handler_errors "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/errors"
	segmentDTO "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/segment"
	userDTO "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/user"
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
	log logger.Logger) *handler {

	h := &handler{
		uc:    uc,
		log:   log,
		group: group,
	}
	h.MapUserRoutes()

	return h
}

// @Summary CreateSegment
// @Tags create
// @Description create segment
// @ID create-segment
// @Accept json
// @Produce json
// @Param input body segmentDTO.SegmentAddInput true "segment with slug, percent(optional)"
// @Success 201 {object} response.Response
// @Failure 400,409,422 {object} response.ErrResponse
// @Failure 500 {object} response.ErrResponse
// @Router /api/v1/segments [post]
func (r *handler) create(c echo.Context) (err error) {
	var (
		now = time.Now()

		// context
		timeout     = 5 * time.Second
		ctx, cancel = context.WithTimeout(c.Request().Context(), timeout)

		// request body dto
		input segmentDTO.SegmentAddInput
	)

	defer cancel()

	err = BindSegmentAddInput(c, &input)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrResponse{
			Message: handler_errors.ErrNotDecodeJSONData.Error(),
		})
	}

	err = ValidateSegmentAddInput(c, &input)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.ErrResponse{
			Message: err.Error(),
		})
	}

	var process string
	if input.Percent != 0 {
		process = segmentDTO.CreateAutoProcess
	} else {
		process = segmentDTO.CreateProcess
	}

	operation := segmentDTO.Operation{
		Segment:     input.Slug,
		OperationAt: now,
		Operation:   process,
		Percent:     input.Percent,
	}

	_, err = r.uc.Create(ctx, operation)
	if err != nil {
		if errors.Is(err, usecase_errors.ErrDB) {
			return c.JSON(http.StatusInternalServerError, response.ErrResponse{
				Message: "InternalServerError",
			})
		}
		if errors.Is(err, usecase_errors.ToFewUsers) {
			return c.JSON(http.StatusBadRequest, response.ErrResponse{
				Message: err.Error(),
			})
		}
		return c.JSON(http.StatusConflict, response.ErrResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, response.Response{
		Message: "success",
	})
}

// @Summary deleteSegment
// @Tags delete
// @Description delete segment
// @ID delete-segment
// @Accept json
// @Produce json
// @Param input body segmentDTO.SegmentDelInput true "segment with slug"
// @Success 200 {object} response.Response
// @Failure 400,404,409,422 {object} response.ErrResponse
// @Failure 500 {object} response.ErrResponse
// @Router /api/v1/segments [delete]
func (r *handler) delete(c echo.Context) (err error) {
	var (
		now = time.Now()
		// context
		timeout     = 5 * time.Second
		ctx, cancel = context.WithTimeout(c.Request().Context(), timeout)

		// request body dto
		input segmentDTO.SegmentDelInput
	)

	defer cancel()

	err = BindSegmentDelInput(c, &input)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrResponse{
			Message: handler_errors.ErrNotDecodeJSONData.Error(),
		})
	}
	err = ValidateSegmentDelInput(c, &input)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.ErrResponse{
			Message: err.Error(),
		})
	}

	operation := userDTO.SegmentTx{
		Slug:      input.Slug,
		CreatedAt: now,
		Operation: segmentDTO.DeleteProcess,
	}

	err = r.uc.Delete(ctx, operation)
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

	return c.JSON(http.StatusOK, response.Response{
		Message: "success",
	})
}
