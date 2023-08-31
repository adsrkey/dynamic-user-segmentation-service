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
	log logger.Logger) *handler {

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

// @Summary addToSegment
// @Tags add
// @Description add user to segment
// @ID add-user-segment
// @Accept json
// @Produce json
// @Param input body dto.AddToSegmentInput true "segment with user_id, slugs_add, slugs_del and ttl (optional)"
// @Success 201 {object} response.Response
// @Failure 400,404,409,422 {object} response.ErrResponse
// @Failure 500 {object} response.ErrResponse
// @Router /api/v1/users/segments [post]
func (h *handler) addToSegment(c echo.Context) (err error) {
	var (
		now = time.Now()
		// context
		timeout     = 5 * time.Second
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
		return c.JSON(http.StatusUnprocessableEntity, response.ErrResponse{
			Message: routeerrs.ErrNotDecodeJSONData.Error(),
		})
	}
	input.OperationAt = now

	if input.Ttl != "" {
		ttl, err := time.Parse(time.RFC3339, input.Ttl)
		if err != nil {
			return c.JSON(http.StatusConflict, response.ErrResponse{
				Message: "err with parse ttl: " + input.Ttl,
			})
		}

		if ttl.Second() > now.Second() {
			return c.JSON(http.StatusConflict, response.ErrResponse{
				Message: "ttl will be greater than now date: " + input.Ttl,
			})
		}
		input.TTL = ttl
	}

	var (
		isSlugsAddEmpty = len(input.SlugsAdd) == 0
		isSlugsDelEmpty = len(input.SlugsDel) == 0

		dup = make([]string, 0, 1)
	)

	// Duplicates. Checking if there are duplicates
	if !isSlugsAddEmpty && !isSlugsDelEmpty {
		for _, v := range input.SlugsDel {
			if contains(input.SlugsAdd, v) {
				dup = append(dup, v)
			}
		}
	}

	var isDupEmpty = len(dup) == 0
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

	return c.JSON(http.StatusCreated, response.Response{
		Message: "success",
	})
}

// @Summary getActiveSegments
// @Tags active segments
// @Description get users active segments
// @ID get-user-segments
// @Accept json
// @Produce json
// @Param input body dto.GetActiveSegments true "get active segments with user_id"
// @Success 200 {object} dto.GetActiveSegmentsResponse
// @Failure 400,404,422 {object} response.ErrResponse
// @Failure 500 {object} response.ErrResponse
// @Router /api/v1/users/segments [get]
func (h *handler) getActiveSegments(c echo.Context) (err error) {
	var (
		// context
		timeout     = 5 * time.Second
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
		return c.JSON(http.StatusUnprocessableEntity, response.ErrResponse{
			Message: routeerrs.ErrNotDecodeJSONData.Error(),
		})
	}

	var (
		slugs []string
	)

	slugs, err = h.uc.GetActiveSegments(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, repoerrs.ErrDB) {
			return c.JSON(http.StatusInternalServerError, response.ErrResponse{
				Message: "InternalServerError",
			})
		}
	}

	if len(slugs) == 0 {
		return c.JSON(http.StatusNotFound, response.ErrResponse{
			Message: "no active segments",
		})
	}

	return c.JSON(http.StatusOK, dto.GetActiveSegmentsResponse{
		UserID: input.UserID,
		Slugs:  slugs,
	})
}

// @Summary reports
// @Tags reports
// @Description get reports
// @ID get-reports
// @Accept json
// @Produce json
// @Param input body dto.ReportInput true "get reports with user_id, year, month"
// @Success 201 {object} response.ReportResponse
// @Failure 400,404,422 {object} response.ErrResponse
// @Failure 500 {object} response.ErrResponse
// @Router /api/v1/users/segments/reports [post]
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
		return c.JSON(http.StatusUnprocessableEntity, response.ErrResponse{
			Message: routeerrs.ErrNotDecodeJSONData.Error(),
		})
	}

	reports, err := h.uc.Reports(ctx, input)
	if err != nil {
		return err
	}
	if len(reports) == 0 {
		return c.JSON(http.StatusNotFound, response.ErrResponse{
			Message: "no reports",
		})
	}

	link, err := linkgenerator.GenerateReportsLink(reports, "http://"+c.Echo().Server.Addr)
	if err != nil {
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

	c.Stream(http.StatusOK, "text/csv", file)
	return
}
