package dto

type Response struct {
	Message string `json:"message"`
}

type ErrResponse struct {
	Message string `json:"message"`
}
