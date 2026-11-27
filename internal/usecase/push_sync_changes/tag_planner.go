package pushsyncchanges

import (
	"github.com/bryanfks-dev/logos/pkg/entity"
	"go.uber.org/zap"
)

type tagBatchPlanner struct {
	idx               *tagIndex
	toCreate          []*entity.Tag
	toUpdate          []*entity.Tag
	pendingCreateIdx  map[*entity.Tag]int
	pendingUpdateByID map[string]int
	log               *zap.Logger
}

func (p *tagBatchPlanner) planTag(tag *entity.Tag) {
	if tag == nil {
		return
	}

	existing := p.idx.findMatch(tag.ID, tag.Name)
	if existing == nil {
		p.pendingCreateIdx[tag] = len(p.toCreate)
		p.toCreate = append(p.toCreate, tag)
		p.idx.register(tag)

		return
	}

	if tag.ID != "" && tag.ID != existing.ID {
		p.idx.idAlias[tag.ID] = existing.ID
	}

	if tag.UpdatedAt.Before(existing.UpdatedAt) {
		p.log.Debug(
			"skipped stale tag update",
			zap.String("tag_id", existing.ID),
		)

		return
	}

	updated := *tag
	updated.ID = existing.ID

	if updated.CreatedAt.IsZero() && !existing.CreatedAt.IsZero() {
		updated.CreatedAt = existing.CreatedAt
	}

	if updated.Name == "" {
		updated.Name = existing.Name
	}

	// Avoid colliding with another tag that already owns updated.Name.
	if owner, ok := p.idx.byName[updated.Name]; ok && owner.ID != existing.ID {
		updated.Name = existing.Name
	}

	if existing.Name != "" && existing.Name != updated.Name {
		delete(p.idx.byName, existing.Name)
	}

	if createIdx, isPendingCreate := p.pendingCreateIdx[existing]; isPendingCreate {
		*existing = updated
		p.toCreate[createIdx] = existing
		p.idx.register(existing)

		return
	}

	if updateIdx, exists := p.pendingUpdateByID[existing.ID]; exists {
		p.toUpdate[updateIdx] = &updated
	} else {
		p.pendingUpdateByID[existing.ID] = len(p.toUpdate)
		p.toUpdate = append(p.toUpdate, &updated)
	}

	p.idx.register(&updated)
}

func (p *tagBatchPlanner) planNoteTags(note *entity.SermonNote) {
	if note == nil || len(note.Tags) == 0 {
		return
	}

	for _, noteTag := range note.Tags {
		if p.idx.findMatch(noteTag.ID, noteTag.Name) != nil {
			continue
		}

		if noteTag.Name == "" {
			p.log.Debug(
				"skipped resolving note tag with empty name",
				zap.String("tag_id", noteTag.ID),
			)

			continue
		}

		created := &entity.Tag{
			ID:   noteTag.ID,
			Name: noteTag.Name,
		}
		p.pendingCreateIdx[created] = len(p.toCreate)
		p.toCreate = append(p.toCreate, created)
		p.idx.register(created)
	}
}
