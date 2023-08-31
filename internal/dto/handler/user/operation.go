package dto

import (
	"time"

	"github.com/google/uuid"
)

type Operation struct {
	UserID      uuid.UUID
	Segment     string
	Operation   string
	OperationAt time.Time
}
