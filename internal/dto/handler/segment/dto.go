package dto

import (
	"encoding/json"
	"io"
	"time"

	"github.com/google/uuid"
)

const (
	CreateProcess     string = "create"
	CreateAutoProcess string = "create_auto"
	DeleteProcess            = "delete"
	OutboxProcess            = "outbox"
)

type SegmentAddInput struct {
	Slug    string `json:"slug" validate:"required"`
	Percent int    `json:"percent"`
}

type SegmentDelInput struct {
	Slug string `json:"slug" validate:"required"`
}

type Operation struct {
	UserID      uuid.UUID
	Segment     string
	Operation   string
	OperationAt time.Time
	Percent     int
	SegmentID   uuid.UUID
}

type Total struct {
	UserIDs    []uuid.UUID
	TotalCount int
}

func (s *SegmentAddInput) FromJSON(r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(s)
}
