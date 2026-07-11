package adapter

import (
	"context"
	"testing"

	"digital.vasic.helixstream/pkg/loginsm"
	"digital.vasic.helixstream/pkg/playback"
)

func TestStubAdapterMetadata(t *testing.T) {
	cases := []struct {
		name        string
		kind        Kind
		country     string
		wantCountry string
	}{
		{"svc-video", Video, "US", "US"},
		{"svc-music", Music, "", CountryAny}, // empty normalises to CountryAny
	}
	for _, c := range cases {
		a := NewStub(c.name, c.kind, c.country, nil, nil, nil)
		if a.Name() != c.name || a.Kind() != c.kind || a.RequiredCountry() != c.wantCountry {
			t.Fatalf("metadata mismatch: %s/%s/%s", a.Name(), a.Kind(), a.RequiredCountry())
		}
	}
}

// The three sub-contracts must be non-nil and usable through the composed adapter.
func TestStubAdapterSubContractsUsable(t *testing.T) {
	a := NewStub("svc", Video, "any",
		loginsm.NewStub(), nil,
		playback.NewStub([]playback.Track{{Kind: playback.Audio, ID: "hi", Rank: 9}}),
	)
	ctx := context.Background()

	if a.Login().Current() != loginsm.StateStart {
		t.Fatal("login sub-contract not at START")
	}
	if _, err := a.Catalog().HomeRows(ctx); err != nil {
		t.Fatalf("catalog sub-contract unusable: %v", err)
	}
	got, err := a.Playback().SelectHighest(ctx, playback.Audio)
	if err != nil || got.ID != "hi" {
		t.Fatalf("playback sub-contract: %+v err=%v", got, err)
	}
}

// Interface satisfaction is enforced at compile time by the var _ assertion in
// adapter.go; this test documents it and guards against a future signature drift.
func TestStubSatisfiesServiceAdapter(t *testing.T) {
	var _ ServiceAdapter = NewStub("x", Video, "any", nil, nil, nil)
}
