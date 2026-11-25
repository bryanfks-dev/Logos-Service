package response

import "time"

type TagResponse struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	UsageCount int       `json:"usage_count,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
