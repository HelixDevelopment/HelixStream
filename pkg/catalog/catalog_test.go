package catalog

import (
	"context"
	"testing"
)

func newFixture() *StubCatalog {
	return NewStub(
		[]string{"Continue Watching", "Popular", "New"},
		[]Tile{
			{ID: "t1", Title: "Alpha Movie", Row: "Popular"},
			{ID: "t2", Title: "Beta Series", Row: "New"},
			{ID: "t3", Title: "Alpha Sequel", Row: "Popular"},
		},
	)
}

func TestHomeRows(t *testing.T) {
	got, err := newFixture().HomeRows(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[0] != "Continue Watching" {
		t.Fatalf("unexpected rows: %v", got)
	}
}

func TestSearch(t *testing.T) {
	c := newFixture()
	cases := []struct {
		query string
		want  int
	}{
		{"alpha", 2}, // case-insensitive, 2 matches
		{"Beta", 1},
		{"zzz", 0},
		{"", 3}, // empty query matches all
	}
	for _, tc := range cases {
		got, err := c.Search(context.Background(), tc.query)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != tc.want {
			t.Fatalf("Search(%q)=%d results, want %d", tc.query, len(got), tc.want)
		}
	}
}

func TestOpen(t *testing.T) {
	c := newFixture()
	if err := c.Open(context.Background(), Tile{ID: "t1"}); err != nil {
		t.Fatalf("Open(t1) unexpected error: %v", err)
	}
	if err := c.Open(context.Background(), Tile{ID: "missing"}); err != ErrNotFound {
		t.Fatalf("Open(missing) = %v, want ErrNotFound", err)
	}
}
