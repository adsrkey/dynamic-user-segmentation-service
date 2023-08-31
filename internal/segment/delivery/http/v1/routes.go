package segment

func (h *handler) MapUserRoutes() {
	h.group.POST("", h.create)
	h.group.DELETE("", h.delete)
}
