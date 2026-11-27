package pushsyncchanges

import (
	"github.com/bryanfks-dev/logos/infra/db"
	"github.com/bryanfks-dev/logos/internal/repository/gathering"
	"github.com/bryanfks-dev/logos/internal/repository/preacher"
	sermonnote "github.com/bryanfks-dev/logos/internal/repository/sermon_note"
	sermonseries "github.com/bryanfks-dev/logos/internal/repository/sermon_series"
	tagrepo "github.com/bryanfks-dev/logos/internal/repository/tag"
	"go.uber.org/zap"
)

type PushSyncChangesUsecase struct {
	txManager              db.TxManager
	sermonNoteRepository   sermonnote.SermonNoteRepository
	gatheringRepository    gathering.GatheringRepository
	sermonSeriesRepository sermonseries.SermonSeriesRepository
	preacherRepository     preacher.PreacherRepository
	tagRepository          tagrepo.TagRepository
	log                    *zap.Logger
}

type PushSyncChangesUsecaseOption struct {
	TxManager              db.TxManager
	SermonNoteRepository   sermonnote.SermonNoteRepository
	GatheringRepository    gathering.GatheringRepository
	SermonSeriesRepository sermonseries.SermonSeriesRepository
	PreacherRepository     preacher.PreacherRepository
	TagRepository          tagrepo.TagRepository
	Log                    *zap.Logger
}

func NewPushSyncChangesUsecase(
	opt *PushSyncChangesUsecaseOption,
) *PushSyncChangesUsecase {
	return &PushSyncChangesUsecase{
		txManager:              opt.TxManager,
		sermonNoteRepository:   opt.SermonNoteRepository,
		gatheringRepository:    opt.GatheringRepository,
		sermonSeriesRepository: opt.SermonSeriesRepository,
		preacherRepository:     opt.PreacherRepository,
		tagRepository:          opt.TagRepository,
		log:                    opt.Log,
	}
}
