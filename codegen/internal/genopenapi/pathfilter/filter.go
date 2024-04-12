// Package pathfilter provides a trie-like structure that allows marking fields that are
// selected or contain children that are selected.
//
// pathfilter is different than a trie because it does NOT have a terminal state.
package pathfilter

import (
	"fmt"
	"strings"

	"github.com/meshapi/grpc-rest-gateway/codegen/internal/fqn"
)

type Instance struct {
	children map[string]*Instance
	Excluded bool
}

func New() *Instance {
	return &Instance{}
}

func (i *Instance) PutString(key string) {
	i.Put(fqn.Parse(&key))
}

func (i *Instance) Put(key fqn.Instance) {
	cursor := i
	for index := 0; index < key.Len(); index++ {
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
	return i.Has(fqn.Parse(&key))
}

func (i *Instance) Has(key fqn.Instance) (bool, *Instance) {
	if i.children == nil {
		return false, nil
	}

	cursor := i
	for index := 0; index < key.Len(); index++ {
		instance, ok := cursor.children[key.Index(index)]
		if !ok {
			return false, nil
		}

		cursor = instance
	}

	return true, cursor
}

func (i *Instance) String() string {
	builder := &strings.Builder{}
	fmt.Fprintf(builder, "Excluded: %v - ", i.Excluded)
	for k, i2 := range i.children {
		fmt.Fprintf(builder, "%s -> [%s]", k, i2)
	}
	return builder.String()
}
