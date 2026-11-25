package converter

import (
	"time"

	"github.com/bryanfks-dev/logos/internal/model"
	"github.com/bryanfks-dev/logos/internal/model/request"
	"github.com/bryanfks-dev/logos/internal/model/response"
	"github.com/bryanfks-dev/logos/pkg/entity"
)

func SyncPushChangesRequestToSyncChangesBatch(
	req *request.SyncPushChangesRequest,
) *model.SyncChangesBatch {
	if req == nil {
		return &model.SyncChangesBatch{
			SermonNotes:  make([]*entity.SermonNote, 0),
			Gatherings:   make([]*entity.Gathering, 0),
			SermonSeries: make([]*entity.SermonSeries, 0),
			Preachers:    make([]*entity.Preacher, 0),
			Tags:         make([]*entity.Tag, 0),
		}
	}

	notes := make([]*entity.SermonNote, 0, len(req.SermonNotes))
	for i := range req.SermonNotes {
		noteItem := &req.SermonNotes[i]
		n := SyncNoteChangesItemToEntity(noteItem)
		if noteItem.UpdatedAt != nil {
			n.UpdatedAt = *noteItem.UpdatedAt
		}
		notes = append(notes, &n)
	}

	gatherings := make([]*entity.Gathering, 0, len(req.Gatherings))
	for i := range req.Gatherings {
		gatheringItem := &req.Gatherings[i]
		g := &entity.Gathering{
			ID:        gatheringItem.ID,
			Name:      gatheringItem.Name,
			ColorHex:  gatheringItem.ColorHex,
			SortOrder: gatheringItem.SortOrder,
		}
		if gatheringItem.CreatedAt != nil {
			g.CreatedAt = *gatheringItem.CreatedAt
		}
		if gatheringItem.UpdatedAt != nil {
			g.UpdatedAt = *gatheringItem.UpdatedAt
		}
		gatherings = append(gatherings, g)
	}

	seriesList := make([]*entity.SermonSeries, 0, len(req.SermonSeries))
	for i := range req.SermonSeries {
		seriesItem := &req.SermonSeries[i]
		s := &entity.SermonSeries{
			ID:          seriesItem.ID,
			Name:        seriesItem.Name,
			Description: seriesItem.Description,
		}
		if seriesItem.CreatedAt != nil {
			s.CreatedAt = *seriesItem.CreatedAt
		}
		if seriesItem.UpdatedAt != nil {
			s.UpdatedAt = *seriesItem.UpdatedAt
		}
		seriesList = append(seriesList, s)
	}

	preachers := make([]*entity.Preacher, 0, len(req.Preachers))
	for i := range req.Preachers {
		preacherItem := &req.Preachers[i]
		p := &entity.Preacher{
			ID:         preacherItem.ID,
			Name:       preacherItem.Name,
			ChurchName: preacherItem.ChurchName,
		}
		if preacherItem.CreatedAt != nil {
			p.CreatedAt = *preacherItem.CreatedAt
		}
		if preacherItem.UpdatedAt != nil {
			p.UpdatedAt = *preacherItem.UpdatedAt
		}
		preachers = append(preachers, p)
	}

	tags := make([]*entity.Tag, 0, len(req.Tags))
	for i := range req.Tags {
		tagItem := &req.Tags[i]
		t := &entity.Tag{
			ID:   tagItem.ID,
			Name: tagItem.Name,
		}
		if tagItem.CreatedAt != nil {
			t.CreatedAt = *tagItem.CreatedAt
		}
		if tagItem.UpdatedAt != nil {
			t.UpdatedAt = *tagItem.UpdatedAt
		}
		tags = append(tags, t)
	}

	return &model.SyncChangesBatch{
		SermonNotes:  notes,
		Gatherings:   gatherings,
		SermonSeries: seriesList,
		Preachers:    preachers,
		Tags:         tags,
	}
}

func SyncPushChangesResultToResponse(
	result *model.SyncPushChangesResult,
) *response.SyncPushChangesResponse {
	if result == nil {
		return &response.SyncPushChangesResponse{
			SyncedAt:          time.Time{},
			MissingContentIDs: make([]string, 0),
			Conflicts:         make([]string, 0),
		}
	}

	return &response.SyncPushChangesResponse{
		SyncedAt:          result.SyncedAt,
		MissingContentIDs: result.MissingContentIDs,
		Conflicts:         result.Conflicts,
	}
}

func SyncPullChangesResultToResponse(
	result *model.SyncPullChangesResult,
) *response.SyncPullChangesResponse {
	if result == nil {
		return &response.SyncPullChangesResponse{
			ServerTime: time.Time{},
			Notes:      make([]response.SermonNoteResponse, 0),
			Gatherings: make([]response.GatheringResponse, 0),
			Series:     make([]response.SermonSeriesResponse, 0),
			Preachers:  make([]response.PreacherResponse, 0),
			Tags:       make([]response.TagResponse, 0),
		}
	}

	var serverTime time.Time
	if !result.ServerTime.IsZero() {
		serverTime = result.ServerTime.UTC()
	}

	notes := make([]response.SermonNoteResponse, 0, len(result.SermonNotes))
	for _, n := range result.SermonNotes {
		if res := SermonNoteToResponse(n); res != nil {
			notes = append(notes, *res)
		}
	}

	gatherings := make([]response.GatheringResponse, 0, len(result.Gatherings))
	for _, g := range result.Gatherings {
		if res := GatheringToResponse(g); res != nil {
			gatherings = append(gatherings, *res)
		}
	}

	series := make([]response.SermonSeriesResponse, 0, len(result.SermonSeries))
	for _, s := range result.SermonSeries {
		if res := SermonSeriesToResponse(s); res != nil {
			series = append(series, *res)
		}
	}

	preachers := make([]response.PreacherResponse, 0, len(result.Preachers))
	for _, p := range result.Preachers {
		if res := PreacherToResponse(p); res != nil {
			preachers = append(preachers, *res)
		}
	}

	tags := make([]response.TagResponse, 0, len(result.Tags))
	for _, t := range result.Tags {
		if res := TagToResponse(t); res != nil {
			tags = append(tags, *res)
		}
	}

	return &response.SyncPullChangesResponse{
		ServerTime: serverTime,
		Notes:      notes,
		Gatherings: gatherings,
		Series:     series,
		Preachers:  preachers,
		Tags:       tags,
	}
}
