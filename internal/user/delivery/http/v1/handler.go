package user

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	response "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler"
	routeerrs "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/errors"
	dto "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/user"
	repoerrs "github.com/adsrkey/dynamic-user-segmentation-service/internal/repository/postgres/errors"
	"github.com/adsrkey/dynamic-user-segmentation-service/internal/usecases"
	"github.com/adsrkey/dynamic-user-segmentation-service/pkg/logger"
	linkgenerator "github.com/adsrkey/dynamic-user-segmentation-service/pkg/utils/link_generator"
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

func Success(c echo.Context) error {
	return c.JSON(http.StatusCreated, response.Response{
		Message: "success",
	})
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
func (h *handler) addToSegment(c echo.Context) (err error) {
	var (
		now = time.Now()
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
	input.OperationAt = now

	// input.OperationAt = c.Request().

	var (
		isSlugsAddEmpty = len(input.SlugsAdd) == 0
		isSlugsDelEmpty = len(input.SlugsDel) == 0

		// slice of duplicates
		dup        = make([]string, 0, 1)
		isDupEmpty = len(dup) == 0
	)

	// Duplicate. Checking if there are duplicates
	if !isSlugsAddEmpty && !isSlugsDelEmpty {
		for _, v := range input.SlugsDel {
			if contains(input.SlugsAdd, v) {
				dup = append(dup, v)
			}
		}
	}

	if !isDupEmpty {
		return c.JSON(http.StatusConflict, response.ErrResponse{
			Message: "contains duplicates data in arrays with duplicate slugs: " + strings.Join(dup[:], ","),
		})
	}

	// select and insert if not exists
	err = h.uc.CreateUser(ctx, input.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, response.ErrResponse{
			Message: err.Error(),
		})
		return
	}

	err = h.uc.AddOrDeleteUserSegment(ctx, input)
	if err != nil {
		c.JSON(http.StatusConflict, response.ErrResponse{
			Message: err.Error(),
		})
		return
	}

	time.Sleep(1 * time.Minute)

	c.JSON(http.StatusInternalServerError, response.Response{
		Message: "success",
	})
	return
}

func (h *handler) getActiveSegments(c echo.Context) (err error) {
	var (
		// context
		timeout     = 1 * time.Second
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

	slugs, err = h.uc.GetActiveSegments(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, repoerrs.ErrDB) {
			return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Message: "InternalServerError",
			})
		}
	}

	if len(slugs) == 0 {
		return c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Message: "no active segments",
		})
	}

	time.Sleep(1 * time.Minute)

	return c.JSON(http.StatusOK, dto.GetActiveSegmentsResponse{
		UserID: input.UserID,
		Slugs:  slugs,
	})
}

// TODO:
func (h *handler) reports(c echo.Context) (err error) {
	var (
		// context
		timeout     = 1 * time.Minute
		ctx, cancel = context.WithTimeout(c.Request().Context(), timeout)

		input dto.ReportInput
	)

	defer cancel()

	// Bind
	err = BindReportInput(c, &input)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrResponse{
			Message: routeerrs.ErrNotDecodeJSONData.Error(),
		})
	}

	// Validate
	err = ValidateReportInput(c, &input)
	if err != nil {
		return err
	}

	// TODO:

	reports, err := h.uc.Reports(ctx, input)
	if err != nil {
		// TODO:
		return err
	}

	link, err := linkgenerator.GenerateReportsLink(reports, "http://"+c.Echo().Server.Addr)
	if err != nil {
		// TODO:
		return err
	}

	return c.JSON(http.StatusCreated, response.ReportResponse{
		Link: link,
	})
}

func (h *handler) file(c echo.Context) (err error) {
	var (
		fileId string
	)

	// TODO:

	fileId = c.QueryParam("file_id")
	if fileId == "" {
		return fmt.Errorf("file_id is empty")
	}

	file, err := os.Open("./files/" + fileId + ".csv")
	defer file.Close()

	if err != nil {
		c.JSON(http.StatusNotFound, map[string]string{
			"message": "file not found",
		})
		return err
	}

	c.Response().Status = http.StatusOK
	http.ServeFile(c.Response().Writer, c.Request(), file.Name())
	return
}
