package user

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	response "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler"
	routeerrs "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/errors"
	dto "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/user"
	usecase_errors "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/usecase/errors"
	"github.com/adsrkey/dynamic-user-segmentation-service/internal/usecases"
	"github.com/adsrkey/dynamic-user-segmentation-service/pkg/logger"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type handler struct {
	uc    usecases.User
	log   logger.Logger
	group *echo.Group
}

func New(
	group *echo.Group,
	uc usecases.User,
	log logger.Logger) *handler { // TODO: to interface

	h := &handler{
		uc:    uc,
		log:   log,
		group: group,
	}
	h.MapUserRoutes()

	return h
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
func (r *handler) addToSegment(c echo.Context) (err error) {
	time.Sleep(10 * time.Second)
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
		return c.JSON(http.StatusBadRequest, response.ErrResponse{
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
		return c.JSON(http.StatusConflict, response.ErrResponse{
			Message: "contains duplicates data in arrays with duplicate slugs: " + strings.Join(dup[:], ","),
		})
	}

	var (
		// just bool variable to check is ErrDB error
		isErrDB bool

		// strings to collect and give it to response
		msgDel string
		msgAdd string

		errDelCh chan struct{}
		errAddCh chan struct{}
	)

	errDelCh = make(chan struct{}, 1)
	errAddCh = make(chan struct{}, 1)

	process := &dto.Process{
		ErrDelCh: errDelCh,
		ErrAddCh: errAddCh,
	}

	wg := &sync.WaitGroup{}

	if !isSlugsAddEmpty {
		wg.Add(1)

		go func(p *dto.Process) {
			defer wg.Done()
			p.ErrAdd = r.uc.AddToSegment(ctx, input, p)
		}(process)

	} else {
		close(process.ErrAddCh)
	}

	// TODO: отменить добавление, если удалить не получилось!!!
	if !isSlugsDelEmpty {
		wg.Add(1)

		go func(p *dto.Process) {
			defer wg.Done()
			p.ErrDel = r.uc.DeleteFromSegment(ctx, input, p)
		}(process)

	} else {
		close(process.ErrDelCh)
	}

	wg.Wait()

	if process.ErrDel != nil {
		if errors.Is(process.ErrDel, usecase_errors.ErrDB) {
			isErrDB = true
		} else {
			msgDel = process.ErrDel.Error()
		}
	}

	if process.ErrAdd != nil {
		if errors.Is(process.ErrAdd, usecase_errors.ErrDB) {
			isErrDB = true
		} else {
			msgAdd = process.ErrAdd.Error()
		}
	}

	if process.ErrDel == nil && process.ErrAdd == nil {
		Success(c)
		return
	}

	if isErrDB {
		c.JSON(http.StatusInternalServerError, response.ErrResponse{
			Message: "InternalServerError",
		})
		return
	}

	response := response.ErrResponse{}

	if msgAdd != "" {
		response.Message = "add: " + msgAdd
	}

	if msgDel != "" {
		response.Message = response.Message + "; " + "delete: " + msgDel
	}

	c.JSON(http.StatusNotFound, response)

	return
}

func Success(c echo.Context) error {
	return c.JSON(http.StatusCreated, response.Response{
		Message: "success",
	})
}

func (r *handler) getActiveSegments(c echo.Context) (err error) {
	time.Sleep(5 * time.Second)
	var (
		// context
		timeout     = 5 * time.Minute
		ctx, cancel = context.WithTimeout(c.Request().Context(), timeout)

		input dto.GetActiveSegments
	)

	defer cancel()

	// Bind
	err = BindGetActiveSegmentsInput(c, &input)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrResponse{
			Message: routeerrs.ErrNotDecodeJSONData.Error(),
		})
	}

	// Validate
	err = ValidateGetActiveSegmentsInput(c, &input)
	if err != nil {
		return err
	}

	var (
		slugs []string
	)

	slugs, err = r.uc.GetActiveSegments(ctx, input.UserID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusNotFound, dto.GetActiveSegmentsResponse{
		UserID: input.UserID,
		Slugs:  slugs,
	})
}
