// Package fqn contains tools to efficiently parse and operate on fully qualified names of format a.b.c or .a.b.c
package dotpath

import (
	"strings"
)

// Instance holds a dotpath string view.
type Instance struct {
	ref   *string
	parts []int
}

// IsAbsolute indicates whether this path is an absolute path. Absolute paths start with '.'
func (i Instance) IsAbsolute() bool {
	return len(i.parts) > 0 && i.parts[0] == 0
}

// Index returns path segment at index.
func (i Instance) Index(index int) string {
	start := 0
	if index > 0 {
		start = i.parts[index-1] + 1
	}

	if index > len(i.parts)-1 {
		return (*i.ref)[start:]
	}

	return (*i.ref)[start:i.parts[index]]
}

// Parts returns all segments, akin to strings.Split(input, ".") but it differs in implementation.
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

// PartsAtDepth returns parts at a specific depth.
// At depth d, the leftmost d segments are excluded.
//
// For instance:
//
// If the path is "a.b.c", the parts at depth 1 are ["b", "c"].
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

// MaxDepth returns the maximum depth this path has.
//
// Note that this depth is the largest depth index, not the count.
func (i Instance) MaxDepth() int {
	return len(i.parts)
}

// NumberOfSegments returns the number of segments in the dot-separated qualified name.
func (i Instance) NumberOfSegments() int {
	return len(i.parts) + 1
}

// StringAtDepth returns the string at a specific depth from the right side.
// For instance, a depth of 0 for "a.b.c" returns "c" and a depth of 2 returns "a.b.c".
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
// no updates to the input string can happen after calling Parse for the lifetime of this object.
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

// ParseString is similar to Parse but it safely copies the string which is less efficient but is safe.
func ParseString(source string) Instance {
	return Parse(&source)
}
