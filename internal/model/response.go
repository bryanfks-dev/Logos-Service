package model

import "time"

type ApiResponse[T any] struct {
	Success bool      `json:"success"`
	Meta    ApiMeta   `json:"meta"`
	Error   *ApiError `json:"error,omitempty"`
	Data    *T        `json:"data,omitempty"`
}

type ApiMeta struct {
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`

	/**
	 * Pagination properties
	 */
	Page      int   `json:"page,omitempty"`
	Size      int   `json:"size,omitempty"`
	TotalPage int64 `json:"total_page,omitempty"`
	TotalItem int64 `json:"total_item,omitempty"`
}

type ApiError struct {
	Code   string       `json:"code"`
	Fields []FieldError `json:"fields,omitempty"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
