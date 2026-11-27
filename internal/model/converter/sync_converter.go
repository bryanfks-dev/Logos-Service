package converter

import (
	"time"

	"github.com/bryanfks-dev/logos/internal/model/request"
	"github.com/bryanfks-dev/logos/internal/model/response"
	pullsyncchanges "github.com/bryanfks-dev/logos/internal/usecase/pull_sync_changes"
	pushsyncchanges "github.com/bryanfks-dev/logos/internal/usecase/push_sync_changes"
)

func SyncPushChangesRequestToPushSyncChangesInput(
	req *request.SyncPushChangesRequest,
) *pushsyncchanges.PushSyncChangesInput {
	if req == nil {
		return &pushsyncchanges.PushSyncChangesInput{
			SermonNotes:  make([]*pushsyncchanges.PushSermonNoteInput, 0),
			Gatherings:   make([]*pushsyncchanges.PushGatheringInput, 0),
			SermonSeries: make([]*pushsyncchanges.PushSermonSeriesInput, 0),
			Preachers:    make([]*pushsyncchanges.PushPreacherInput, 0),
			Tags:         make([]*pushsyncchanges.PushTagInput, 0),
		}
	}

	notes := make(
		[]*pushsyncchanges.PushSermonNoteInput,
		0,
		len(req.SermonNotes),
	)
	for i := range req.SermonNotes {
		noteItem := &req.SermonNotes[i]
		n := SyncNoteChangesItemToPushSermonNoteInput(noteItem)
		notes = append(notes, n)
	}

	gatherings := make(
		[]*pushsyncchanges.PushGatheringInput,
		0,
		len(req.Gatherings),
	)
	for i := range req.Gatherings {
		gatheringItem := &req.Gatherings[i]
		g := &pushsyncchanges.PushGatheringInput{
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

	seriesList := make(
		[]*pushsyncchanges.PushSermonSeriesInput,
		0,
		len(req.SermonSeries),
	)
	for i := range req.SermonSeries {
		seriesItem := &req.SermonSeries[i]
		s := &pushsyncchanges.PushSermonSeriesInput{
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

	preachers := make(
		[]*pushsyncchanges.PushPreacherInput,
		0,
		len(req.Preachers),
	)
	for i := range req.Preachers {
		preacherItem := &req.Preachers[i]
		p := &pushsyncchanges.PushPreacherInput{
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

	tags := make([]*pushsyncchanges.PushTagInput, 0, len(req.Tags))
	for i := range req.Tags {
		tagItem := &req.Tags[i]
		t := &pushsyncchanges.PushTagInput{
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

	return &pushsyncchanges.PushSyncChangesInput{
		SermonNotes:  notes,
		Gatherings:   gatherings,
		SermonSeries: seriesList,
		Preachers:    preachers,
		Tags:         tags,
	}
}

func PushSyncChangesResultToResponse(
	result *pushsyncchanges.PushSyncChangesResult,
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
	result *pullsyncchanges.SyncPullChangesResult,
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
