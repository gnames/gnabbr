package abbrname_test

import (
	"testing"

	"github.com/gnames/gnabbr/pkg/abbrname"
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

func TestExpand(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("Carex alba", abbrname.Expand("C. alba", "Carex"))
	assert.Equal("Carex alba var. nigra L.",
		abbrname.Expand("C.  alba var. nigra L.", "Carex"))
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

func TestReadings(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		verbatim, name string
		res            []string
	}{
		{"(CS. ambigua", "Cs. ambigua",
			[]string{"S. ambigua", "C. ambigua", "Cs. ambigua"}},
		{"LL. clethroides,", "Ll. clethroides",
			[]string{"L. clethroides", "Ll. clethroides"}},
		{"CI. moniliferum.", "Ci. moniliferum",
			[]string{"I. moniliferum", "C. moniliferum", "Cl. moniliferum", "Ci. moniliferum"}},
		{"JHa. alba var. nigra", "Jha. alba var. nigra",
			[]string{"Ha. alba var. nigra", "J. alba var. nigra", "Jha. alba var. nigra"}},
		{"Cs. ambigua", "Cs. ambigua", nil},
		{"C. ambigua", "C. ambigua", nil},
		{"CSA. ambigua", "Csa. ambigua", nil},
		{"CS. ambigua", "Carex ambigua", nil},
		{"", "Cs. ambigua", nil},
	}
	for _, v := range tests {
		assert.Equal(v.res, abbrname.Readings(v.verbatim, v.name), v.verbatim)
	}
}
