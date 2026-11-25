package model

import (
	"time"

	"github.com/bryanfks-dev/logos/pkg/entity"
)

type SyncChangesBatch struct {
	SermonNotes  []*entity.SermonNote
	Gatherings   []*entity.Gathering
	SermonSeries []*entity.SermonSeries
	Preachers    []*entity.Preacher
	Tags         []*entity.Tag
}

type SyncPushChangesResult struct {
	SyncedAt          time.Time
	MissingContentIDs []string
	Conflicts         []string
}

type SyncPullChangesResult struct {
	ServerTime   time.Time
	SermonNotes  []*entity.SermonNote
	Gatherings   []*entity.Gathering
	SermonSeries []*entity.SermonSeries
	Preachers    []*entity.Preacher
	Tags         []*entity.Tag
}
