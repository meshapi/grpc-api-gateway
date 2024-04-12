package internal

type DependencyKind uint8

const (
	DependencyKindEnum DependencyKind = iota
	DependencyKindMessage
)

type SchemaDependency struct {
	FQN  string
	Kind DependencyKind
}

func (s SchemaDependency) IsSet() bool {
	return s.FQN != ""
}

type schemaDependencyNode struct {
	Kind  DependencyKind
	Count uint16
}

type SchemaDependencyStore map[string]schemaDependencyNode

func AddDependency(store SchemaDependencyStore, dependency SchemaDependency) SchemaDependencyStore {
	if store == nil {
		store = SchemaDependencyStore{}
	}
	store.Add(dependency)
	return store
}

func AddDependencies(destinationStore, sourceStore SchemaDependencyStore) SchemaDependencyStore {
	if destinationStore == nil {
		return sourceStore
	}
	for key, value := range sourceStore {
		if item, exists := destinationStore[key]; exists {
			item.Count += value.Count
		} else {
			destinationStore[key] = value
		}
	}
	return destinationStore
}

func (s SchemaDependencyStore) Add(dependency SchemaDependency) {
	if item, exists := s[dependency.FQN]; exists {
		item.Count++
		s[dependency.FQN] = item
	} else {
		s[dependency.FQN] = schemaDependencyNode{
			Kind:  dependency.Kind,
			Count: 1,
		}
	}
}

func (s SchemaDependencyStore) Drop(fqn string) {
	if item, exists := s[fqn]; exists {
		item.Count--
		if item.Count == 0 {
			delete(s, fqn)
		} else {
			s[fqn] = item
		}
	}
}

func (s SchemaDependencyStore) HasAny() bool {
	return len(s) > 0
}

func (s SchemaDependencyStore) Copy() SchemaDependencyStore {
	newStore := make(SchemaDependencyStore, len(s))
	for key, value := range s {
		newStore[key] = value
	}
	return newStore
}
