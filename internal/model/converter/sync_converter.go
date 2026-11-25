package converter

import (
	"time"

	"github.com/bryanfks-dev/logos/internal/model"
	"github.com/bryanfks-dev/logos/pkg/entity"
	"github.com/bryanfks-dev/logos/pkg/util"
)

func SermonNoteToSyncItem(note *entity.SermonNote) model.SyncNoteItem {
	if note == nil {
		return model.SyncNoteItem{}
	}

	tags := make([]string, 0, len(note.Tags))
	for _, tag := range note.Tags {
		tags = append(tags, tag.Name)
	}

	verses := make([]model.SyncNoteVerse, 0, len(note.Verses))
	for _, v := range note.Verses {
		verses = append(verses, model.SyncNoteVerse{
			ID:          v.ID,
			Translation: v.Translation,
			BookID:      v.BookID,
			Chapter:     v.Chapter,
			VerseStart:  v.VerseStart,
			VerseEnd:    v.VerseEnd,
			DisplayRef:  v.DisplayRef,
		})
	}

	isDeleted := note.DeletedAt.Valid
	updatedAt := note.UpdatedAt

	return model.SyncNoteItem{
		ID:          note.ID,
		GatheringID: note.GatheringID,
		SeriesID:    note.SeriesID,
		PreacherID:  note.PreacherID,
		Title:       note.Title,
		SermonDate:  note.SermonDate,
		ContentHash: note.ContentHash,
		PlainText:   note.PlainText,
		ContentJSON: note.ContentJSON,
		Version:     note.Version,
		Tags:        tags,
		Verses:      verses,
		IsDeleted:   isDeleted,
		UpdatedAt:   &updatedAt,
	}
}

func SyncItemToSermonNote(item *model.SyncNoteItem) entity.SermonNote {
	if item == nil {
		return entity.SermonNote{}
	}

	version := item.Version
	if version < 1 {
		version = 1
	}

	verses := make([]entity.NoteVerse, 0, len(item.Verses))
	for _, v := range item.Verses {
		verses = append(verses, entity.NoteVerse{
			ID:          v.ID,
			NoteID:      item.ID,
			Translation: v.Translation,
			BookID:      v.BookID,
			Chapter:     v.Chapter,
			VerseStart:  v.VerseStart,
			VerseEnd:    v.VerseEnd,
			DisplayRef:  v.DisplayRef,
		})
	}

	tags := make([]entity.Tag, 0, len(item.Tags))
	for _, tagName := range item.Tags {
		tags = append(tags, entity.Tag{
			Name: tagName,
		})
	}

	return entity.SermonNote{
		ID:          item.ID,
		GatheringID: item.GatheringID,
		SeriesID:    item.SeriesID,
		PreacherID:  item.PreacherID,
		Title:       item.Title,
		SermonDate:  item.SermonDate,
		ContentHash: item.ContentHash,
		PlainText:   item.PlainText,
		ContentJSON: item.ContentJSON,
		Version:     version,
		Verses:      verses,
		Tags:        tags,
	}
}

func SyncPushRequestToBatch(req *model.SyncPushRequest) *model.SyncBatch {
	if req == nil {
		return &model.SyncBatch{
			Notes:        make([]*entity.SermonNote, 0),
			SermonNotes:  make([]*entity.SermonNote, 0),
			Gatherings:   make([]*entity.Gathering, 0),
			Series:       make([]*entity.SermonSeries, 0),
			SermonSeries: make([]*entity.SermonSeries, 0),
			Preachers:    make([]*entity.Preacher, 0),
			Tags:         make([]*entity.Tag, 0),
		}
	}

	rawNotes := util.CoalesceSlice(req.Notes, req.SermonNotes)

	notes := make([]*entity.SermonNote, 0, len(rawNotes))
	for i := range rawNotes {
		item := &rawNotes[i]
		n := SyncItemToSermonNote(item)
		if item.UpdatedAt != nil {
			n.UpdatedAt = *item.UpdatedAt
		}
		notes = append(notes, &n)
	}

	gatherings := make([]*entity.Gathering, 0, len(req.Gatherings))
	for i := range req.Gatherings {
		item := &req.Gatherings[i]
		g := &entity.Gathering{
			ID:        item.ID,
			Name:      item.Name,
			ColorHex:  item.ColorHex,
			SortOrder: item.SortOrder,
		}
		if item.CreatedAt != nil {
			g.CreatedAt = *item.CreatedAt
		}
		if item.UpdatedAt != nil {
			g.UpdatedAt = *item.UpdatedAt
		}
		gatherings = append(gatherings, g)
	}

	rawSeries := util.CoalesceSlice(req.Series, req.SermonSeries)

	seriesList := make([]*entity.SermonSeries, 0, len(rawSeries))
	for i := range rawSeries {
		item := &rawSeries[i]
		s := &entity.SermonSeries{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
		}
		if item.CreatedAt != nil {
			s.CreatedAt = *item.CreatedAt
		}
		if item.UpdatedAt != nil {
			s.UpdatedAt = *item.UpdatedAt
		}
		seriesList = append(seriesList, s)
	}

	preachers := make([]*entity.Preacher, 0, len(req.Preachers))
	for i := range req.Preachers {
		item := &req.Preachers[i]
		p := &entity.Preacher{
			ID:         item.ID,
			Name:       item.Name,
			ChurchName: item.ChurchName,
		}
		if item.CreatedAt != nil {
			p.CreatedAt = *item.CreatedAt
		}
		if item.UpdatedAt != nil {
			p.UpdatedAt = *item.UpdatedAt
		}
		preachers = append(preachers, p)
	}

	tags := make([]*entity.Tag, 0, len(req.Tags))
	for i := range req.Tags {
		item := &req.Tags[i]
		t := &entity.Tag{
			ID:   item.ID,
			Name: item.Name,
		}
		if item.CreatedAt != nil {
			t.CreatedAt = *item.CreatedAt
		}
		if item.UpdatedAt != nil {
			t.UpdatedAt = *item.UpdatedAt
		}
		tags = append(tags, t)
	}

	return &model.SyncBatch{
		Notes:        notes,
		SermonNotes:  notes,
		Gatherings:   gatherings,
		Series:       seriesList,
		SermonSeries: seriesList,
		Preachers:    preachers,
		Tags:         tags,
	}
}

func SyncPushResultToResponse(
	result *model.SyncPushResult,
) *model.SyncPushResponse {
	if result == nil {
		return &model.SyncPushResponse{
			SyncedAt:          time.Time{},
			MissingContentIDs: make([]string, 0),
			Conflicts:         make([]string, 0),
		}
	}

	missing := util.NonNilSlice(result.MissingContentIDs)
	conflicts := util.NonNilSlice(result.Conflicts)

	return &model.SyncPushResponse{
		SyncedAt:          result.SyncedAt,
		MissingContentIDs: missing,
		Conflicts:         conflicts,
	}
}

func SermonNoteToResponse(note *entity.SermonNote) *model.SermonNoteResponse {
	if note == nil {
		return nil
	}

	tags := make([]string, 0, len(note.Tags))
	for _, tag := range note.Tags {
		tags = append(tags, tag.Name)
	}

	verses := make([]model.SyncNoteVerse, 0, len(note.Verses))
	for _, v := range note.Verses {
		verses = append(verses, model.SyncNoteVerse{
			ID:          v.ID,
			Translation: v.Translation,
			BookID:      v.BookID,
			Chapter:     v.Chapter,
			VerseStart:  v.VerseStart,
			VerseEnd:    v.VerseEnd,
			DisplayRef:  v.DisplayRef,
		})
	}

	var deletedAt *time.Time
	isDeleted := note.DeletedAt.Valid
	if isDeleted {
		deletedAt = &note.DeletedAt.Time
	}

	return &model.SermonNoteResponse{
		ID:          note.ID,
		GatheringID: note.GatheringID,
		SeriesID:    note.SeriesID,
		PreacherID:  note.PreacherID,
		Title:       note.Title,
		SermonDate:  note.SermonDate,
		FilePath:    note.FilePath,
		ContentHash: note.ContentHash,
		PlainText:   note.PlainText,
		ContentJSON: note.ContentJSON,
		Version:     note.Version,
		SyncStatus:  note.SyncStatus,
		Tags:        tags,
		Verses:      verses,
		CreatedAt:   note.CreatedAt,
		UpdatedAt:   note.UpdatedAt,
		DeletedAt:   deletedAt,
		IsDeleted:   isDeleted,
	}
}

func GatheringToResponse(gathering *entity.Gathering) *model.GatheringResponse {
	if gathering == nil {
		return nil
	}

	var deletedAt *time.Time
	isDeleted := gathering.DeletedAt.Valid
	if isDeleted {
		deletedAt = &gathering.DeletedAt.Time
	}

	return &model.GatheringResponse{
		ID:        gathering.ID,
		Name:      gathering.Name,
		ColorHex:  gathering.ColorHex,
		SortOrder: gathering.SortOrder,
		CreatedAt: gathering.CreatedAt,
		UpdatedAt: gathering.UpdatedAt,
		DeletedAt: deletedAt,
		IsDeleted: isDeleted,
	}
}

func SermonSeriesToResponse(
	series *entity.SermonSeries,
) *model.SermonSeriesResponse {
	if series == nil {
		return nil
	}

	var deletedAt *time.Time
	isDeleted := series.DeletedAt.Valid
	if isDeleted {
		deletedAt = &series.DeletedAt.Time
	}

	return &model.SermonSeriesResponse{
		ID:          series.ID,
		Name:        series.Name,
		Description: series.Description,
		CreatedAt:   series.CreatedAt,
		UpdatedAt:   series.UpdatedAt,
		DeletedAt:   deletedAt,
		IsDeleted:   isDeleted,
	}
}

func PreacherToResponse(preacher *entity.Preacher) *model.PreacherResponse {
	if preacher == nil {
		return nil
	}

	var deletedAt *time.Time
	isDeleted := preacher.DeletedAt.Valid
	if isDeleted {
		deletedAt = &preacher.DeletedAt.Time
	}

	return &model.PreacherResponse{
		ID:         preacher.ID,
		Name:       preacher.Name,
		ChurchName: preacher.ChurchName,
		CreatedAt:  preacher.CreatedAt,
		UpdatedAt:  preacher.UpdatedAt,
		DeletedAt:  deletedAt,
		IsDeleted:  isDeleted,
	}
}

func TagToResponse(tag *entity.Tag) *model.TagResponse {
	if tag == nil {
		return nil
	}

	return &model.TagResponse{
		ID:         tag.ID,
		Name:       tag.Name,
		UsageCount: tag.UsageCount,
		CreatedAt:  tag.CreatedAt,
		UpdatedAt:  tag.UpdatedAt,
	}
}

func SyncPullResultToResponse(
	result *model.SyncPullResult,
) *model.SyncPullResponse {
	if result == nil {
		return &model.SyncPullResponse{
			ServerTime: time.Time{},
			Notes:      make([]*model.SermonNoteResponse, 0),
			Gatherings: make([]*model.GatheringResponse, 0),
			Series:     make([]*model.SermonSeriesResponse, 0),
			Preachers:  make([]*model.PreacherResponse, 0),
			Tags:       make([]*model.TagResponse, 0),
		}
	}

	var serverTime time.Time
	if !result.ServerTime.IsZero() {
		serverTime = result.ServerTime.UTC()
	}

	rawNotes := util.CoalesceSlice(result.Notes, result.SermonNotes)

	notes := make([]*model.SermonNoteResponse, 0, len(rawNotes))
	for _, n := range rawNotes {
		if res := SermonNoteToResponse(n); res != nil {
			notes = append(notes, res)
		}
	}

	gatherings := make([]*model.GatheringResponse, 0, len(result.Gatherings))
	for _, g := range result.Gatherings {
		if res := GatheringToResponse(g); res != nil {
			gatherings = append(gatherings, res)
		}
	}

	rawSeries := util.CoalesceSlice(result.Series, result.SermonSeries)

	series := make([]*model.SermonSeriesResponse, 0, len(rawSeries))
	for _, s := range rawSeries {
		if res := SermonSeriesToResponse(s); res != nil {
			series = append(series, res)
		}
	}

	preachers := make([]*model.PreacherResponse, 0, len(result.Preachers))
	for _, p := range result.Preachers {
		if res := PreacherToResponse(p); res != nil {
			preachers = append(preachers, res)
		}
	}

	tags := make([]*model.TagResponse, 0, len(result.Tags))
	for _, t := range result.Tags {
		if res := TagToResponse(t); res != nil {
			tags = append(tags, res)
		}
	}

	return &model.SyncPullResponse{
		ServerTime: serverTime,
		Notes:      notes,
		Gatherings: gatherings,
		Series:     series,
		Preachers:  preachers,
		Tags:       tags,
	}
}
