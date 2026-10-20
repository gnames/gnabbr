package hints_test

import (
	"testing"

	"github.com/gnames/gnabbr/pkg/doc"
	"github.com/gnames/gnabbr/pkg/hints"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	assert := assert.New(t)
	names := []doc.Name{
		{Name: "Cyperus fuscus", Start: 0},   // 0
		{Name: "Coleoptera", Start: 50},      // 1, uninomial
		{Name: "Carex alba", Start: 100},     // 2
		{Name: "Juncus effusus", Start: 150}, // 3, other prefix
		{Name: "C. nigra", Start: 200},       // 4
		{Name: "Carex rubra", Start: 250},    // 5
		{Name: "C. alba", Start: 300},        // 6
	}
	scope := []int{0, 1, 2, 3, 4, 5, 6}

	hs := hints.New(names, scope, 1000, 5)
	assert.Equal([]string{"Carex", "Coleoptera", "Cyperus"}, hs[4])
	assert.Equal([]string{"Carex", "Coleoptera", "Cyperus"}, hs[6])

	hs = hints.New(names, scope, 160, 2)
	assert.Equal([]string{"Carex", "Coleoptera"}, hs[4])
	assert.Equal([]string{"Carex"}, hs[6])

	// Names outside the scope give no hints.
	hs = hints.New(names, []int{3, 4}, 1000, 5)
	assert.Empty(hs[4])
}

func TestMerge(t *testing.T) {
	assert := assert.New(t)
	res := hints.Merge(5,
		[]string{"Carex", "Cyperus"},
		[]string{"Cladium", "Carex", "Cuscuta"},
	)
	assert.Equal([]string{"Carex", "Cladium", "Cyperus", "Cuscuta"}, res)
	assert.Equal([]string{"Carex"}, hints.Merge(1, []string{"Carex", "Cyperus"}))
	assert.Empty(hints.Merge(5))
}

func TestExtend(t *testing.T) {
	assert := assert.New(t)
	group := []string{"Carex", "Cladium", "Cyperus"}
	assert.Equal([]string{"Cyperus", "Carex", "Cladium"},
		hints.Extend([]string{"Cyperus"}, group, 5))
	assert.Equal([]string{"Cyperus", "Carex"},
		hints.Extend([]string{"Cyperus"}, group, 2))
	assert.Equal(group, hints.Extend(nil, group, 5))
}
