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
	names := []doc.Name{
		{Name: "Carex alba", Start: 0},  // 0, not abbreviated
		{Name: "C. alba", Start: 20},    // 1
		{Name: "C. nigra", Start: 300},  // 2
		{Name: "Ca. nigra", Start: 310}, // 3, other prefix
		{Name: "C. rubra", Start: 2000}, // 4, after a gap
		{Name: "C. alba", Start: 10},    // 5, before 1
		{Name: "C. viridis", Start: 2500},
		{Name: "C. viridis", Start: 3100}, // 7, span limit
	}
	scope := []int{0, 5, 1, 2, 3, 4, 6, 7}
	gs := group.New(names, scope, 2, 1000, 1000)
	require.Len(t, gs, 4)

	assert.Equal("C", gs[0].Prefix)
	assert.Equal(2, gs[0].Section)
	assert.Equal([]int{5, 1, 2}, gs[0].Members)
	assert.Equal(10, gs[0].Start)
	assert.Equal(300, gs[0].End)

	assert.Equal([]int{4, 6}, gs[1].Members)
	assert.Equal([]int{7}, gs[2].Members)

	assert.Equal("Ca", gs[3].Prefix)
	assert.Equal([]int{3}, gs[3].Members)
}
