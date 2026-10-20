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
	assert.Equal([]string{"Carex alba", "Coleoptera", "Cyperus fuscus"}, hs[4])
	// The nearest name of each genus, and names with the epithet.
	assert.Equal([]string{"Carex rubra", "Carex alba", "Coleoptera", "Cyperus fuscus"}, hs[6])

	hs = hints.New(names, scope, 160, 2)
	assert.Equal([]string{"Carex alba", "Coleoptera"}, hs[4])
	assert.Equal([]string{"Carex rubra"}, hs[6])

	// The limit counts genera: a name with the epithet is still added.
	hs = hints.New(names, scope, 1000, 1)
	assert.Equal([]string{"Carex rubra", "Carex alba"}, hs[6])

	// Names outside the scope give no hints.
	hs = hints.New(names, []int{3, 4}, 1000, 5)
	assert.Empty(hs[4])
}

func TestGenera(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("Carex", hints.Genus("Carex alba"))
	assert.Equal("Carex", hints.Genus("Carex"))
	assert.Equal([]string{"Carex", "Cyperus"},
		hints.Genera([]string{"Carex alba", "Carex", "Cyperus fuscus", "Carex nigra"}))
	assert.Empty(hints.Genera(nil))
}

func TestMerge(t *testing.T) {
	assert := assert.New(t)
	res := hints.Merge(5,
		[]string{"Carex", "Cyperus"},
		[]string{"Cladium", "Carex", "Cuscuta"},
	)
	assert.Equal([]string{"Carex", "Cladium", "Cyperus", "Cuscuta"}, res)
	assert.Equal([]string{"Carex"}, hints.Merge(1, []string{"Carex", "Cyperus"}))
	// The limit counts genera, not names.
	assert.Equal([]string{"Carex alba", "Carex nigra", "Cyperus fuscus"},
		hints.Merge(2, []string{"Carex alba", "Carex nigra", "Cyperus fuscus", "Cladium x"}))
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
	assert.Equal([]string{"Carex alba", "Carex nigra"},
		hints.Extend([]string{"Carex alba"}, []string{"Carex nigra", "Cyperus fuscus"}, 1))
}
