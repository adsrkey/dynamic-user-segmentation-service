package segment

import "github.com/adsrkey/dynamic-user-segmentation-service/pkg/middleware"

func (h *handler) MapUserRoutes() {
	h.group.Use(middleware.ValidContentType)
	h.group.POST("", h.create)
	h.group.DELETE("", h.delete)
}
