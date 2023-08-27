package user

import (
	"context"
	"net/http"
	"strings"
	"time"

	routeerrs "github.com/adsrkey/dynamic-user-segmentation-service/internal/delivery/http/v1/routes/errors"
	"github.com/adsrkey/dynamic-user-segmentation-service/internal/delivery/http/v1/routes/user/dto"
	usecase_errors "github.com/adsrkey/dynamic-user-segmentation-service/internal/usecase/errors"
	uc "github.com/adsrkey/dynamic-user-segmentation-service/internal/usecase/user"
	"github.com/adsrkey/dynamic-user-segmentation-service/pkg/logger"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type route struct {
	uc  *uc.UseCase
	log logger.Logger
}

func New(g *echo.Group, uc *uc.UseCase, log logger.Logger) { // TODO: to interface
	route := &route{
		uc:  uc,
		log: log,
	}

	g.POST("", route.addToSegment)
	g.GET("/:user_id/segments", route.getActiveSegments)
}

func contains(elems []string, v string) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}

// ВОПРОС: Если добавляются и удаляются те же самые слаги, то что делать, удалять сразу у пользователя?
// получается мы добавляем 3 слага и удаляем потом опять же их
func (r *route) addToSegment(c echo.Context) (err error) {
	var (
		// context
		timeout     = 5 * time.Minute
		ctx, cancel = context.WithTimeout(c.Request().Context(), timeout)

		// request body dto
		input dto.AddToSegmentInput
	)

	defer cancel()

	// Bind
	err = BindAddToSegmentInput(c, &input)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrResponse{
			Message: routeerrs.ErrNotDecodeJSONData.Error(),
		})
	}

	// Validate
	err = ValidateAddToSegmentInput(c, &input)
	if err != nil {
		return err
	}

	var (
		isSlugsAddEmpty = len(input.SlugsAdd) == 0
		isSlugsDelEmpty = len(input.SlugsDel) == 0

		// slice of duplicates
		dup = make([]string, 0, 1)
	)

	// Duplicate. Checking if there are duplicates
	if !isSlugsAddEmpty && !isSlugsDelEmpty {
		for _, v := range input.SlugsDel {
			if contains(input.SlugsAdd, v) {
				dup = append(dup, v)
			}
		}
	}

	var (
		isDupEmpty = len(dup) == 0
	)

	if !isDupEmpty {
		return c.JSON(http.StatusConflict, dto.ErrResponse{
			Message: "contains duplicates data in arrays with duplicate slugs: " + strings.Join(dup[:], ","),
		})
	}

	err = r.uc.CreateUser(ctx, input.UserID)
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

	var (
		// because after we collect the errors strings and give them in response like "add: ... delete: ..."
		errAdd error
		errDel error

		// just bool variable to check is ErrDB error
		isErrDB bool

		// strings to collect and give it to response
		msgDel string
		msgAdd string
	)

	if !isSlugsAddEmpty {
		errAdd = r.uc.AddToSegment(ctx, input.SlugsAdd, input.UserID)
	}

	if !isSlugsDelEmpty {
		errDel = r.uc.DelFromSegment(ctx, input.SlugsDel, input.UserID)
	}

	if errDel != nil {
		if errors.Is(errDel, usecase_errors.ErrDB) {
			isErrDB = true
		} else {
			msgDel = errDel.Error()
		}
	}

	if errAdd != nil {
		if errors.Is(errAdd, usecase_errors.ErrDB) {
			isErrDB = true
		} else {
			msgAdd = errAdd.Error()
		}
	}

	if isErrDB {
		c.JSON(http.StatusInternalServerError, dto.ErrResponse{
			Message: "InternalServerError",
		})
		return
	}

	if errDel == nil && errAdd == nil {
		return Success(c)
	}

	response := dto.ErrResponse{}
	if msgAdd != "" {
		response.Message = "add: " + msgAdd + ";"
	}
	if msgDel != "" {
		response.Message = response.Message + "delete: " + msgDel + ";"
	}

	c.JSON(http.StatusNotFound, response)

	return
}

func Success(c echo.Context) error {
	return c.JSON(http.StatusCreated, dto.Response{
		Message: "success",
	})
}

func (r *route) getActiveSegments(c echo.Context) (err error) {
	var (
		// context
		timeout     = 5 * time.Minute
		ctx, cancel = context.WithTimeout(c.Request().Context(), timeout)
	)

	defer cancel()

	input := c.Param("user_id")

	var (
		slugs []string
	)

	userID, err := uuid.Parse(input)
	if err != nil {
		return err
	}

	slugs, err = r.uc.GetActiveSegments(ctx, userID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusNotFound, dto.GetActiveSegmentsResponse{
		UserID: userID,
		Slugs:  slugs,
	})
}
