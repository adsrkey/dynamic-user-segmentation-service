package dto

import (
	"time"

	"github.com/google/uuid"
)

const (
	AddProcess    string = "create"
	DeleteProcess        = "delete"
	OutboxProcess        = "outbox"
)

type AddToSegmentInput struct {
	UserID      uuid.UUID `json:"user_id" validate:"required"`
	SlugsAdd    []string  `json:"slugs_add" validate:"required"`
	SlugsDel    []string  `json:"slugs_del" validate:"required"`
	Ttl         string    `json:"ttl,omitempty"`
	OperationAt time.Time
	TTL         time.Time
}

type SegmentTx struct {
	UserID    uuid.UUID
	Slug      string
	Operation string
	CreatedAt time.Time
	TTL       time.Time
	SegmentID uuid.UUID
}

// ttl in format
type TTLTx struct {
	TTL string
}

// ttl return format
type TTLTxR struct {
	UserID    uuid.UUID
	SegmentID uuid.UUID
	Slug      string
}

type GetActiveSegments struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
}

type GetActiveSegmentsResponse struct {
	UserID uuid.UUID `json:"user_id"`
	Slugs  []string  `json:"slugs"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}
