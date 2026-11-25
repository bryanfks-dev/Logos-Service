package converter

import (
	"time"

	"github.com/bryanfks-dev/logos/internal/model/response"
	"github.com/bryanfks-dev/logos/pkg/entity"
)

func PreacherToResponse(preacher *entity.Preacher) *response.PreacherResponse {
	if preacher == nil {
		return nil
	}

	var deletedAt *time.Time
	isDeleted := preacher.DeletedAt.Valid
	if isDeleted {
		deletedAt = &preacher.DeletedAt.Time
	}

	return &response.PreacherResponse{
		ID:         preacher.ID,
		Name:       preacher.Name,
		ChurchName: preacher.ChurchName,
		CreatedAt:  preacher.CreatedAt,
		UpdatedAt:  preacher.UpdatedAt,
		DeletedAt:  deletedAt,
	}
}
