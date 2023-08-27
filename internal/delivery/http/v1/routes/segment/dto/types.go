package dto

import (
	"encoding/json"
	"io"
)

type SlugInput struct {
	Slug string `json:"slug" validate:"required"`
}

type ErrResponse struct {
	Message string `json:"message"`
}

func (s *SlugInput) FromJSON(r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(s)
}

type Response struct {
	Message string `json:"message"`
}