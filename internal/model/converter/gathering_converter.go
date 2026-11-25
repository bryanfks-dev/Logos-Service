package converter

import (
	"time"

	"github.com/bryanfks-dev/logos/internal/model/response"
	"github.com/bryanfks-dev/logos/pkg/entity"
)

func GatheringToResponse(
	gathering *entity.Gathering,
) *response.GatheringResponse {
	if gathering == nil {
		return nil
	}

	var deletedAt *time.Time
	isDeleted := gathering.DeletedAt.Valid
	if isDeleted {
		deletedAt = &gathering.DeletedAt.Time
	}

	return &response.GatheringResponse{
		ID:        gathering.ID,
		Name:      gathering.Name,
		ColorHex:  gathering.ColorHex,
		SortOrder: gathering.SortOrder,
		CreatedAt: gathering.CreatedAt,
		UpdatedAt: gathering.UpdatedAt,
		DeletedAt: deletedAt,
	}
}
