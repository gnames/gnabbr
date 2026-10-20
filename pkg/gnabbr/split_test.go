package gnabbr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRuns(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		msg    string
		genera []string
		strong []bool
		runs   [][]int
	}{
		{"one genus", []string{"A", "A", "A"}, nil, [][]int{{0, 1, 2}}},
		{"one name", []string{"A"}, nil, [][]int{{0}}},
		{"two runs", []string{"A", "A", "B", "B"}, nil, [][]int{{0, 1}, {2, 3}}},
		{"interruption", []string{"A", "A", "B", "A", "A"}, nil, [][]int{{0, 1, 3, 4}}},
		{"return", []string{"A", "A", "B", "B", "A", "A"}, nil, [][]int{{0, 1}, {2, 3}, {4, 5}}},
		{"stray first", []string{"B", "A", "A"}, nil, [][]int{{1, 2}}},
		{"stray last", []string{"A", "A", "B"}, nil, [][]int{{0, 1}}},
		{"alternating", []string{"A", "B", "A", "B"}, nil, nil},
		{"strong interruption", []string{"A", "A", "B", "A"},
			[]bool{false, false, true, false}, [][]int{{0, 1}, {2}}},
		{"strong pair", []string{"A", "B"}, []bool{true, false}, [][]int{{0}}},
		{"empty", nil, nil, nil},
	}
	for _, v := range tests {
		assert.Equal(v.runs, runs(v.genera, v.strong), v.msg)
	}
}
