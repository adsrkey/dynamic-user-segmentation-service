package domain

import "github.com/google/uuid"

type Segment struct {
	ID   uuid.UUID `json:"id"`
	Slug string    `json:"slug"`
}

type User struct {
	ID    uuid.UUID `json:"id"`
	Slugs []string  `json:"slugs"`
}
