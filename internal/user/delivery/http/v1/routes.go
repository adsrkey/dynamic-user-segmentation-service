package user

func (h *handler) MapUserRoutes() {
	h.group.POST("", h.addToSegment)
	// h.group.GET("/:user_id/segments", h.getActiveSegments)
	h.group.GET("/segments", h.getActiveSegments)
	h.group.POST("/reports", h.reports)
	h.group.GET("/files", h.file)
}
