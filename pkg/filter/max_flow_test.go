package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaxFlowExample(t *testing.T) {
	// See example in: testdata/maximum-flow.dot
	g := []vertex{
		{0, []edge{{Dest: 2, Cap: 3}}},
		{1, []edge{{Dest: 2, Cap: 5}, {Dest: 3, Cap: 4}}},
		{2, []edge{{Dest: 6, Cap: 2}}},
		{3, []edge{{Dest: 4, Cap: 2}}},
		{4, []edge{{Dest: 6, Cap: 3}}},
		{5, []edge{{Dest: 0, Cap: 3}, {Dest: 1, Cap: 1}}},
		{6, nil},
	}
	expectedFlow := [][]int{
		{2},
		{0, 1},
		{2},
		{1},
		{1},
		{2, 1},
		{},
	}

	maxFlow(g, 5, 6)

	for i, v := range g {
		for j, e := range v.OutEdges {
			assert.Equal(t, expectedFlow[i][j], e.Flow)
		}
	}
}
