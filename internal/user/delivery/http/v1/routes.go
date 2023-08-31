package user

import "github.com/adsrkey/dynamic-user-segmentation-service/pkg/middleware"

func (h *handler) MapUserRoutes() {
	validJson := h.group.Group("/segments")
	validJson.Use(middleware.ValidContentType)

	validJson.POST("", h.addToSegment)
	// h.group.GET("/:user_id/segments", h.getActiveSegments)

	validJson.GET("", h.getActiveSegments)
	validJson.POST("/reports", h.reports)

	group := validJson.Group("/files")

	group.GET("", h.file)

	h.group = group
}
