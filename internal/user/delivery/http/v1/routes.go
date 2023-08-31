package user

func (h *handler) MapUserRoutes() {
	segments := h.group.Group("/segments")

	segments.POST("", h.addToSegment)
	// h.group.GET("/:user_id/segments", h.getActiveSegments)
	segments.GET("", h.getActiveSegments)
	segments.POST("/reports", h.reports)
	segments.GET("/files", h.file)

	h.group = segments
}
