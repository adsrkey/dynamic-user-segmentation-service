package dto

import (
	"github.com/google/uuid"
)

type AddToSegmentInput struct {
	UserID   uuid.UUID `json:"user_id" validate:"required"`
	SlugsAdd []string  `json:"slugs_add" validate:"required"`
	SlugsDel []string  `json:"slugs_del" validate:"required"`
}

type GetActiveSegments struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
}

type GetActiveSegmentsResponse struct {
	UserID uuid.UUID `json:"user_id"`
	Slugs  []string  `json:"slugs"`
}
