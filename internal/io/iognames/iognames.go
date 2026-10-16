// Package iognames calls the abbr_names endpoint of the gnames REST API.
package iognames

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gnames/gnlib/ent/abbr"
)

// Client calls abbr_names.
type Client struct {
	url  string
	http *http.Client
}

// New creates a client for the gnames API at url, for example
// `https://verifier.globalnames.org/api/v1/`.
func New(url string) *Client {
	return &Client{
		url:  strings.TrimSuffix(url, "/") + "/abbr_names",
		http: &http.Client{Timeout: 2 * time.Minute},
	}
}

// AbbrNames sends a group of names to abbr_names.
func (c *Client) AbbrNames(ctx context.Context, inp abbr.Input) (abbr.Output, error) {
	var res abbr.Output
	body, err := json.Marshal(inp)
	if err != nil {
		return res, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return res, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return res, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return res, fmt.Errorf("abbr_names at %s: %s", c.url, resp.Status)
	}
	err = json.NewDecoder(resp.Body).Decode(&res)
	return res, err
}
