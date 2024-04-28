package trie

import (
	"github.com/meshapi/grpc-rest-gateway/dotpath"
)

type Node struct {
	children map[string]*Node
	terminal bool
}

func New(items ...string) *Node {
	n := &Node{}
	for _, item := range items {
		n.Add(dotpath.Parse(&item))
	}
	return n
}

func (n *Node) Iterate(cb func(string)) {
	if len(n.children) == 0 {
		return
	}

	for key, node := range n.children {
		if node.terminal {
			cb(key)
		}
		node.iterate(key+".", cb)
	}
}

func (n *Node) iterate(prefix string, cb func(string)) {
	if len(n.children) == 0 {
		return
	}

	for key, node := range n.children {
		if node.terminal {
			cb(prefix + key)
		}
		node.iterate(prefix+key+".", cb)
	}
}

func (n *Node) AddString(key string) {
	n.Add(dotpath.Parse(&key))
}

func (n *Node) Add(key dotpath.Instance) {
	cursor := n
	for index := 0; index < key.NumberOfSegments(); index++ {
		key := key.Index(index)
		if cursor.children == nil {
			cursor.children = make(map[string]*Node, 1)
		}
		if node, exists := cursor.children[key]; exists {
			cursor = node
		} else {
			node = &Node{}
			cursor.children[key] = node
			cursor = node
		}
	}
	cursor.terminal = true
}

func (n *Node) HasCommonPrefix(key dotpath.Instance) bool {
	if n.children == nil {
		return false
	}

	cursor := n
	for index := 0; index < key.NumberOfSegments(); index++ {
		node, ok := cursor.children[key.Index(index)]
		if !ok {
			return false
		}

		if node.terminal {
			return true
		}

		cursor = node
	}

	return cursor.terminal
}

func (n *Node) HasCommonPrefixString(key string) bool {
	return n.HasCommonPrefix(dotpath.Parse(&key))
}
