package trie_test

import (
	"strconv"
	"testing"

	"github.com/meshapi/grpc-rest-gateway/trie"
)

func TestTrie(t *testing.T) {
	testCases := []struct {
		Inputs      []string
		ExpectTrue  []string
		ExpectFalse []string
	}{
		{
			Inputs:      []string{"a.b.c", "a.d"},
			ExpectTrue:  []string{"a.b.c", "a.d", "a.d.something_else"},
			ExpectFalse: []string{"a", "a.b", "d", "x"},
		},
	}

	for index, tt := range testCases {
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			node := trie.New(tt.Inputs...)

			for _, value := range tt.ExpectTrue {
				if !node.HasCommonPrefixString(value) {
					t.Logf("expected Has(%q) == true but received false", value)
					t.Fail()
				}
			}

			for _, value := range tt.ExpectFalse {
				if node.HasCommonPrefixString(value) {
					t.Logf("expected Has(%q) == false but received true", value)
					t.Fail()
				}
			}
		})
	}
}
