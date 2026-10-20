package group_test

import (
	"testing"

	"github.com/gnames/gnabbr/pkg/doc"
	"github.com/gnames/gnabbr/pkg/group"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	assert := assert.New(t)
	occs := []doc.Occurrence{
		{Name: "Carex alba", Page: 1},           // 0, not abbreviated
		{Name: "C. alba", Page: 1, Offset: 10},  // 1
		{Name: "C. nigra", Page: 3},             // 2
		{Name: "Ca. nigra", Page: 3, Offset: 5}, // 3, other prefix
		{Name: "C. rubra", Page: 20},            // 4, after a gap
		{Name: "C. alba", Page: 21, Part: "p2"}, // 5, other part
		{Name: "C. alba", Page: 1, Offset: 2},   // 6, before 1
		{Name: "C. viridis", Page: 30},          // 7
		{Name: "C. viridis", Page: 36},          // 8, span limit
	}
	gs := group.New(occs, 10, 12)
	require.Len(t, gs, 5)

	assert.Equal("C", gs[0].Prefix)
	assert.Equal([]int{6, 1, 2}, gs[0].Members)
	assert.Equal(1, gs[0].StartPage)
	assert.Equal(3, gs[0].EndPage)

	assert.Equal([]int{4, 7}, gs[1].Members)
	assert.Equal([]int{8}, gs[2].Members)

	assert.Equal("Ca", gs[3].Prefix)
	assert.Equal([]int{3}, gs[3].Members)

	assert.Equal("p2", gs[4].Part)
	assert.Equal([]int{5}, gs[4].Members)
}
