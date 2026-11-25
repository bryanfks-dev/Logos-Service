package response

import "time"

type PreacherResponse struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	ChurchName string     `json:"church_name"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at"`
}
