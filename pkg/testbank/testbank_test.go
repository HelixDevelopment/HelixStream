package testbank

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"digital.vasic.helixstream/pkg/smoothav"
)

// moduleRoot walks up from the test cwd to the dir containing go.mod.
func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate module root (go.mod)")
		}
		dir = parent
	}
}

// TestLoadShippedBanks proves the ACTUAL shipped bank files parse + validate.
func TestLoadShippedBanks(t *testing.T) {
	root := moduleRoot(t)
	cases := []struct {
		path         string
		wantName     string
		wantServices int
	}{
		{"banks/streaming_playback.yaml", "streaming_playback", 2},
		{"examples/atmosphere/streaming_ru_services.yaml", "atmosphere_streaming_ru_services", 6},
	}
	for _, c := range cases {
		t.Run(c.path, func(t *testing.T) {
			b, err := LoadFile(filepath.Join(root, c.path))
			if err != nil {
				t.Fatalf("LoadFile(%s) rejected a valid bank: %v", c.path, err)
			}
			if b.Name != c.wantName {
				t.Fatalf("name = %q, want %q", b.Name, c.wantName)
			}
			if len(b.Services) != c.wantServices {
				t.Fatalf("services = %d, want %d", len(b.Services), c.wantServices)
			}
			if b.SchemaVersion == "" || b.BankVersion == "" {
				t.Fatal("shipped bank must carry schema_version + bank_version")
			}
		})
	}
}

// The bank's smooth-AV budgets must bridge to the P1 smoothav.Thresholds.
func TestSmoothAVBridgesToP1Contract(t *testing.T) {
	root := moduleRoot(t)
	b, err := LoadFile(filepath.Join(root, "examples/atmosphere/streaming_ru_services.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var th smoothav.Thresholds = b.Services[0].Playback.SmoothAV.ToThresholds()
	if th.FreezeSSIM != 0.99 || th.MaxDropped != 3 {
		t.Fatalf("ToThresholds bridge wrong: %+v", th)
	}
}

const goldenGood = `
schema_version: "1.0"
name: golden_good
bank_version: "1.0"
services:
  - id: svc
    adapter: generic_androidtv
    kind: video
    topology: any
    geo: { required_country: any }
    playback: { transport: [pause, resume, stop], random_seeks: 1 }
`

// goldenBad maps a description → a bank that MUST be rejected, and (optionally)
// the sentinel error it must match.
func goldenBad() []struct {
	name string
	yaml string
	is   error
} {
	base := func(body string) string { return body }
	return []struct {
		name string
		yaml string
		is   error
	}{
		{"missing schema_version", base(`
name: b
bank_version: "1.0"
services: [{id: s, adapter: a, kind: video, topology: any}]
`), ErrMissingSchemaVersion},
		{"unsupported schema_version", base(`
schema_version: "2.0"
name: b
bank_version: "1.0"
services: [{id: s, adapter: a, kind: video, topology: any}]
`), ErrUnsupportedSchemaVersion},
		{"missing bank_version", base(`
schema_version: "1.0"
name: b
services: [{id: s, adapter: a, kind: video, topology: any}]
`), ErrMissingBankVersion},
		{"no services", base(`
schema_version: "1.0"
name: b
bank_version: "1.0"
services: []
`), ErrNoServices},
		{"invalid kind", base(`
schema_version: "1.0"
name: b
bank_version: "1.0"
services: [{id: s, adapter: a, kind: hologram, topology: any}]
`), ErrInvalidKind},
		{"invalid topology", base(`
schema_version: "1.0"
name: b
bank_version: "1.0"
services: [{id: s, adapter: a, kind: video, topology: triple_display}]
`), ErrInvalidTopology},
		{"invalid geo country", base(`
schema_version: "1.0"
name: b
bank_version: "1.0"
services: [{id: s, adapter: a, kind: video, topology: any, geo: {required_country: usa}}]
`), ErrInvalidGeoCountry},
		{"invalid transport verb", base(`
schema_version: "1.0"
name: b
bank_version: "1.0"
services: [{id: s, adapter: a, kind: video, topology: any, playback: {transport: [rewind]}}]
`), ErrInvalidTransport},
		{"missing service id", base(`
schema_version: "1.0"
name: b
bank_version: "1.0"
services: [{adapter: a, kind: video, topology: any}]
`), ErrMissingServiceID},
		{"missing adapter", base(`
schema_version: "1.0"
name: b
bank_version: "1.0"
services: [{id: s, kind: video, topology: any}]
`), ErrMissingAdapter},
		{"unknown/typo field (strict schema)", base(`
schema_version: "1.0"
name: b
bank_version: "1.0"
serrvices: [{id: s, adapter: a, kind: video, topology: any}]
`), nil}, // KnownFields parse error — no sentinel, just must be rejected
		{"not yaml", "::: not : valid : yaml :::", nil},
	}
}

// TestLoaderRejectsMalformedBanks is the anti-bluff negative suite: a loader
// that ACCEPTS any of these is a FAIL (§11.4.6). Each must be rejected.
func TestLoaderRejectsMalformedBanks(t *testing.T) {
	for _, c := range goldenBad() {
		t.Run(c.name, func(t *testing.T) {
			_, err := Load(strings.NewReader(c.yaml))
			if err == nil {
				t.Fatalf("loader ACCEPTED a malformed bank (%s) — this is a §11.4 bluff", c.name)
			}
			if c.is != nil && !errors.Is(err, c.is) {
				t.Fatalf("error = %v, want to match %v", err, c.is)
			}
		})
	}
}

// TestLoaderSelfValidation (§11.4.107(10)): prove the loader is not a bluff — it
// PASSes the golden-good bank AND rejects every golden-bad bank. A loader that
// passed its golden-bad fixtures would be a bluffing validator.
func TestLoaderSelfValidation(t *testing.T) {
	if _, err := Load(strings.NewReader(goldenGood)); err != nil {
		t.Fatalf("self-validation: golden-good bank was rejected: %v", err)
	}
	for _, c := range goldenBad() {
		if _, err := Load(strings.NewReader(c.yaml)); err == nil {
			t.Fatalf("self-validation: golden-bad bank %q was ACCEPTED — the loader is bluffing", c.name)
		}
	}
}
