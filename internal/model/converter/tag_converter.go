package converter

import (
	"github.com/bryanfks-dev/logos/internal/model/response"
	"github.com/bryanfks-dev/logos/pkg/entity"
)

func TagToResponse(tag *entity.Tag) *response.TagResponse {
	if tag == nil {
		return nil
	}

	return &response.TagResponse{
		ID:         tag.ID,
		Name:       tag.Name,
		UsageCount: tag.UsageCount,
		CreatedAt:  tag.CreatedAt,
		UpdatedAt:  tag.UpdatedAt,
	}
}
