package response

import (
	"time"
)

type SyncPushChangesResponse struct {
	SyncedAt          time.Time `json:"synced_at"`
	MissingContentIDs []string  `json:"missing_content_ids"`
	Conflicts         []string  `json:"conflicts"`
}

type SyncPullChangesResponse struct {
	ServerTime time.Time              `json:"server_time"`
	Notes      []SermonNoteResponse   `json:"notes"`
	Gatherings []GatheringResponse    `json:"gatherings"`
	Series     []SermonSeriesResponse `json:"series"`
	Preachers  []PreacherResponse     `json:"preachers"`
	Tags       []TagResponse          `json:"tags"`
}
