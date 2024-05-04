// Package pathfilter provides a trie-like structure that allows marking fields that are
// selected or contain children that are selected.
//
// pathfilter is different than a trie because it does NOT have a terminal state.
package pathfilter

import (
	"github.com/meshapi/grpc-api-gateway/dotpath"
)

type Instance struct {
	children map[string]*Instance
	Excluded bool
}

func New() *Instance {
	return &Instance{}
}

func (i *Instance) PutString(key string) {
	i.Put(dotpath.Parse(&key))
}

func (i *Instance) Put(key dotpath.Instance) {
	cursor := i
	for index := 0; index < key.NumberOfSegments(); index++ {
		key := key.Index(index)
		if cursor.children == nil {
			cursor.children = make(map[string]*Instance, 1)
		}
		if instance, exists := cursor.children[key]; exists {
			cursor = instance
		} else {
			instance = &Instance{}
			cursor.children[key] = instance
			cursor = instance
		}
	}
	cursor.Excluded = true
}

func (i *Instance) HasString(key string) (bool, *Instance) {
	return i.Has(dotpath.Parse(&key))
}

func (i *Instance) Has(key dotpath.Instance) (bool, *Instance) {
	if i.children == nil {
		return false, nil
	}

	cursor := i
	for index := 0; index < key.NumberOfSegments(); index++ {
		instance, ok := cursor.children[key.Index(index)]
		if !ok {
			return false, nil
		}

		cursor = instance
	}

	return true, cursor
}
