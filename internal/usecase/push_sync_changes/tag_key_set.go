package pushsyncchanges

type tagKeySet struct {
	ids       []string
	names     []string
	seenIDs   map[string]struct{}
	seenNames map[string]struct{}
}

func (s *tagKeySet) add(id string, name string) {
	s.addID(id)
	s.addName(name)
}

func (s *tagKeySet) addID(id string) {
	if id == "" {
		return
	}

	if _, ok := s.seenIDs[id]; ok {
		return
	}

	s.seenIDs[id] = struct{}{}
	s.ids = append(s.ids, id)
}

func (s *tagKeySet) addName(name string) {
	if name == "" {
		return
	}

	if _, ok := s.seenNames[name]; ok {
		return
	}

	s.seenNames[name] = struct{}{}
	s.names = append(s.names, name)
}
