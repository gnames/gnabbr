package hints_test

import (
	"testing"

	"github.com/gnames/gnabbr/pkg/doc"
	"github.com/gnames/gnabbr/pkg/hints"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	assert := assert.New(t)
	occs := []doc.Occurrence{
		{Name: "Carex alba", Page: 1},                // 0
		{Name: "Cyperus fuscus", Page: 2},            // 1
		{Name: "Juncus effusus", Page: 2, Offset: 9}, // 2
		{Name: "C. alba", Page: 3},                   // 3
		{Name: "Carex nigra", Page: 4},               // 4
		{Name: "C. rubra", Page: 4, Offset: 7},       // 5
		{Name: "C. viridis", Page: 30},               // 6, out of window
		{Name: "C. alba", Page: 4, Part: "p2"},       // 7, other part
		{Name: "Cy. fuscus", Page: 5},                // 8
	}
	hs := hints.New(occs, 10, 5)
	assert.Equal([]string{"Cyperus", "Carex"}, hs[3])
	assert.Equal([]string{"Carex", "Cyperus"}, hs[5])
	assert.Equal([]string{"Cyperus"}, hs[8])
	assert.NotContains(hs, 6)
	assert.NotContains(hs, 7)
	assert.Len(hints.New(occs, 10, 1)[3], 1)
}
