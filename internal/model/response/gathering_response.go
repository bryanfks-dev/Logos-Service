package response

import "time"

type GatheringResponse struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	ColorHex  string     `json:"color_hex"`
	SortOrder int        `json:"sort_order"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}
