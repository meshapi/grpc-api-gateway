// Package fqn contains tools to efficiently parse and operate on fully qualified names of format .a.b.c
package fqn

import (
	"strings"
)

type Instance struct {
	ref   *string
	parts []int
}

func (i Instance) Parts() []string {
	l := len(i.parts)
	result := make([]string, l+1)
	start := 0
	for j, index := range i.parts {
		result[j] = (*i.ref)[start:index]
		start = index + 1
	}
	result[l] = (*i.ref)[start:]

	return result
}

func (i Instance) PartsAtDepth(d int) []string {
	l := len(i.parts)
	if d == l {
		return i.Parts()
	}

	result := make([]string, d+1)
	start := i.parts[l-d-1] + 1
	counter := 0
	for j := l - d; j < l; j++ {
		index := i.parts[j]
		result[counter] = (*i.ref)[start:index]
		counter++
		start = i.parts[j] + 1
	}
	result[d] = (*i.ref)[start:]

	return result
}

// MaxDepth returns the maximum depth this FQN has. Note that this depth is the largest depth index, not the count.
func (i Instance) MaxDepth() int {
	return len(i.parts)
}

// StringAtDepth returns the string at a certain depth from the right side.
// For instance, depth of 0 for a.b.c is c and depth of 2 is a.b.c
func (i Instance) StringAtDepth(d int) string {
	l := len(i.parts)

	if d == l {
		return *i.ref
	}

	return (*i.ref)[i.parts[l-d-1]+1:]
}

// Parse takes a string pointer, keeps it as a reference.
//
// NOTE: This tool is meant to use pointers in order to avoid copying data, because of this
// no updates to the input string can happen after calling Parse for the duration the lifetime of this object.
func Parse(source *string) Instance {
	parts := make([]int, strings.Count(*source, "."))

	l := len(*source)
	start := 0
	counter := 0
	for start < l {
		index := strings.IndexByte((*source)[start:], '.')
		if index == -1 {
			break
		}
		parts[counter] = start + index
		counter++
		start += index + 1
	}

	return Instance{
		ref:   source,
		parts: parts,
	}
}
