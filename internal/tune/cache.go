package tune

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sync"

	"github.com/gnames/gnabbr/pkg/gnabbr"
	"github.com/gnames/gnlib/ent/abbr"
	_ "modernc.org/sqlite"
)

// cache keeps abbr_names answers in SQLite, keyed by the gnames
// version and the request, so reruns with the same groups and hints
// make no calls.
type cache struct {
	db      *sql.DB
	client  gnabbr.AbbrNamer
	version string

	mu     sync.Mutex
	hits   int
	misses int
}

func newCache(path, version string, client gnabbr.AbbrNamer) (*cache, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(10000)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS abbr_names(
  key    TEXT PRIMARY KEY,  -- sha256 of gnames version and request
  output BLOB NOT NULL      -- abbr_names answer, JSON
)`)
	if err != nil {
		db.Close()
		return nil, err
	}
	return &cache{db: db, client: client, version: version}, nil
}

// AbbrNames implements gnabbr.AbbrNamer.
func (c *cache) AbbrNames(ctx context.Context, inp abbr.Input) (abbr.Output, error) {
	var res abbr.Output
	req, err := json.Marshal(inp)
	if err != nil {
		return res, err
	}
	sum := sha256.Sum256(append([]byte(c.version+"\n"), req...))
	key := hex.EncodeToString(sum[:])

	var b []byte
	err = c.db.QueryRowContext(ctx,
		"SELECT output FROM abbr_names WHERE key = ?", key).Scan(&b)
	switch {
	case err == nil:
		c.count(true)
		err = json.Unmarshal(b, &res)
		return res, err
	case !errors.Is(err, sql.ErrNoRows):
		return res, err
	}

	c.count(false)
	if res, err = c.client.AbbrNames(ctx, inp); err != nil {
		return res, err
	}
	if b, err = json.Marshal(res); err != nil {
		return res, err
	}
	_, err = c.db.ExecContext(ctx,
		"INSERT OR REPLACE INTO abbr_names VALUES (?, ?)", key, b)
	return res, err
}

func (c *cache) count(hit bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if hit {
		c.hits++
	} else {
		c.misses++
	}
}

func (c *cache) Close() error {
	return c.db.Close()
}
