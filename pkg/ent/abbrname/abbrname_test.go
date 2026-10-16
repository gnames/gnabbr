package abbrname_test

import (
	"testing"

	"github.com/gnames/gnabbr/pkg/ent/abbrname"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name, prefix, epithet string
		ok                    bool
	}{
		{"C. elegans", "C", "elegans", true},
		{"Cae. elegans Maupas, 1900", "Cae", "elegans", true},
		{"Ca. elegans var. alba", "Ca", "elegans", true},
		{"Caen. elegans", "", "", false},
		{"C elegans", "", "", false},
		{"c. elegans", "", "", false},
		{"C. Elegans", "", "", false},
		{"CA. elegans", "", "", false},
		{"Caenorhabditis elegans", "", "", false},
		{"C.", "", "", false},
	}
	for _, v := range tests {
		res, ok := abbrname.Parse(v.name)
		assert.Equal(v.ok, ok, v.name)
		assert.Equal(v.prefix, res.Prefix, v.name)
		assert.Equal(v.epithet, res.Epithet, v.name)
	}
}

func TestGenus(t *testing.T) {
	assert := assert.New(t)
	tests := []struct{ name, genus string }{
		{"Carex alba L.", "Carex"},
		{"Carex", "Carex"},
		{"C. alba", ""},
		{"carex alba", ""},
		{"CAREX alba", ""},
		{"", ""},
	}
	for _, v := range tests {
		assert.Equal(v.genus, abbrname.Genus(v.name), v.name)
	}
}
