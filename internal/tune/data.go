package tune

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gnames/gnabbr/pkg/doc"
)

// Item is a sampled BHL item (a row of items.jsonl).
type Item struct {
	ItemID int     `json:"item_id"`
	Split  string  `json:"split"`
	Weight float64 `json:"weight"`
	GroupQ float64 `json:"group_q"`
}

// Label is the label of a sampled abbreviated occurrence (a row of
// labels.jsonl).
type Label struct {
	ItemID       int     `json:"item_id"`
	OccurrenceID int     `json:"occurrence_id"`
	Selected     int     `json:"selected"`
	Genus        *string `json:"label_genus"`
	Source       *string `json:"label_source"`
	Confidence   *string `json:"confidence"`
}

// Kind is the kind of a label: `antecedent high`, `last_genus`,
// `no label`...
func (l Label) Kind() string {
	switch {
	case l.Source == nil:
		return "no label"
	case *l.Source == "antecedent" && l.Confidence != nil:
		return "antecedent " + *l.Confidence
	default:
		return *l.Source
	}
}

// readItems reads items.jsonl.
func readItems(dir string) ([]Item, error) {
	var res []Item
	err := readJSONL(filepath.Join(dir, "items.jsonl"), func(b []byte) error {
		var it Item
		err := json.Unmarshal(b, &it)
		res = append(res, it)
		return err
	})
	return res, err
}

// readLabels reads the selected labels of labels.jsonl by item id.
func readLabels(dir string) (map[int][]Label, error) {
	res := make(map[int][]Label)
	err := readJSONL(filepath.Join(dir, "labels.jsonl"), func(b []byte) error {
		var l Label
		if err := json.Unmarshal(b, &l); err != nil {
			return err
		}
		if l.Selected == 1 {
			res[l.ItemID] = append(res[l.ItemID], l)
		}
		return nil
	})
	return res, err
}

// readDoc reads docs/<item>.json.
func readDoc(dir string, item int) (doc.Document, error) {
	var res doc.Document
	b, err := os.ReadFile(filepath.Join(dir, "docs", strconv.Itoa(item)+".json"))
	if err != nil {
		return res, err
	}
	err = json.Unmarshal(b, &res)
	return res, err
}

func readJSONL(path string, f func([]byte) error) error {
	fh, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fh.Close()
	sc := bufio.NewScanner(fh)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	var n int
	for sc.Scan() {
		n++
		if err = f(sc.Bytes()); err != nil {
			return fmt.Errorf("%s:%d: %w", path, n, err)
		}
	}
	return sc.Err()
}
