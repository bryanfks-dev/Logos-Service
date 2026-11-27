package pushsyncchanges

import "github.com/bryanfks-dev/logos/pkg/entity"

type tagIndex struct {
	byID    map[string]*entity.Tag
	byName  map[string]*entity.Tag
	idAlias map[string]string
}

func (idx *tagIndex) register(tag *entity.Tag) {
	if tag == nil {
		return
	}

	if tag.ID != "" {
		idx.byID[tag.ID] = tag
	}

	if tag.Name != "" {
		idx.byName[tag.Name] = tag
	}
}

func (idx *tagIndex) findMatch(id string, name string) *entity.Tag {
	if id != "" {
		if existing := idx.findByID(id); existing != nil {
			return existing
		}
	}

	if name != "" {
		return idx.byName[name]
	}

	return nil
}

func (idx *tagIndex) findByID(id string) *entity.Tag {
	if aliasedID, ok := idx.idAlias[id]; ok {
		if existing, found := idx.byID[aliasedID]; found {
			return existing
		}
	}

	return idx.byID[id]
}
