package dto

import (
	"time"

	"github.com/google/uuid"
)

type Report struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Segment     string
	Operation   string
	OperationAt time.Time
}

type ReportInput struct {
	UserID uuid.UUID  `json:"user_id" validate:"required"`
	Year   int        `json:"year" validate:"required"`
	Month  time.Month `json:"month" validate:"required"`
}
