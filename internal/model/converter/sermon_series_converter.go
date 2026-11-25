package converter

import (
	"time"

	"github.com/bryanfks-dev/logos/internal/model/response"
	"github.com/bryanfks-dev/logos/pkg/entity"
)

func SermonSeriesToResponse(
	series *entity.SermonSeries,
) *response.SermonSeriesResponse {
	if series == nil {
		return nil
	}

	var deletedAt *time.Time
	isDeleted := series.DeletedAt.Valid
	if isDeleted {
		deletedAt = &series.DeletedAt.Time
	}

	return &response.SermonSeriesResponse{
		ID:          series.ID,
		Name:        series.Name,
		Description: series.Description,
		CreatedAt:   series.CreatedAt,
		UpdatedAt:   series.UpdatedAt,
		DeletedAt:   deletedAt,
	}
}
