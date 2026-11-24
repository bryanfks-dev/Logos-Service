package model

import "time"

type APIResponse[T any] struct {
	Success bool      `json:"success"`
	Data    *T        `json:"data,omitempty"`
	Meta    APIMeta   `json:"meta"`
	Error   *APIError `json:"error,omitempty"`
}

type APIMeta struct {
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`
}

type APIError struct {
	Code   string       `json:"code"`
	Errors []FieldError `json:"errors,omitempty"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
