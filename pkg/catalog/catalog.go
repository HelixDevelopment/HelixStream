// Package catalog defines the service-neutral catalog-browse contract
// (DESIGN.md §7): home rows, tiles, search, detail-page open. The engine drives
// it via vision navigation on non-introspectable TV-Compose UIs (§11.4.117);
// the adapter supplies per-service anchors.
//
// P1 NOTE (skeleton): the interface + an in-memory stub only. No device nav.
package catalog

import (
	"context"
	"errors"
)

// Tile is one catalog entry (a show / album / video card).
type Tile struct {
	ID    string
	Title string
	Row   string // the home-screen row it appeared in ("" if from search)
}

// ErrNotFound is returned when Open is asked for a tile the catalog does not know.
var ErrNotFound = errors.New("catalog: tile not found")

// CatalogMap is the core catalog-browse contract (one of the three
// ServiceAdapter sub-contracts).
type CatalogMap interface {
	// HomeRows returns the discovered home-screen row labels.
	HomeRows(ctx context.Context) ([]string, error)
	// Search returns tiles matching a query (case-insensitive substring in the
	// stub; real adapters drive the service search UI).
	Search(ctx context.Context, query string) ([]Tile, error)
	// Open navigates to a tile's detail page. Returns ErrNotFound if unknown.
	Open(ctx context.Context, tile Tile) error
}

// StubCatalog is an in-memory reference implementation for contract tests.
type StubCatalog struct {
	rows  []string
	tiles []Tile
}

// NewStub builds a StubCatalog seeded with the given rows and tiles.
func NewStub(rows []string, tiles []Tile) *StubCatalog {
	return &StubCatalog{rows: rows, tiles: tiles}
}

func (c *StubCatalog) HomeRows(context.Context) ([]string, error) {
	return append([]string(nil), c.rows...), nil
}

func (c *StubCatalog) Search(_ context.Context, query string) ([]Tile, error) {
	var out []Tile
	for _, t := range c.tiles {
		if containsFold(t.Title, query) {
			out = append(out, t)
		}
	}
	return out, nil
}

func (c *StubCatalog) Open(_ context.Context, tile Tile) error {
	for _, t := range c.tiles {
		if t.ID == tile.ID {
			return nil
		}
	}
	return ErrNotFound
}

// containsFold is a tiny case-insensitive substring check (ASCII-fold; adequate
// for the stub — real adapters do not use it).
func containsFold(s, sub string) bool {
	if sub == "" {
		return true
	}
	ls, lsub := toLowerASCII(s), toLowerASCII(sub)
	for i := 0; i+len(lsub) <= len(ls); i++ {
		if ls[i:i+len(lsub)] == lsub {
			return true
		}
	}
	return false
}

func toLowerASCII(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		}
	}
	return string(b)
}
