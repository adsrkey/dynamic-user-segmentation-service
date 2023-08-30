package dto

import (
	"encoding/json"
	"io"
)

type SlugInput struct {
	Slug string `json:"slug" validate:"required"`
}

func (s *SlugInput) FromJSON(r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(s)
}

const (
	CreateProcess     string = "create"
	CreateAutoProcess string = "create_auto"
	DeleteProcess            = "delete"
	OutboxProcess            = "outbox"
)
