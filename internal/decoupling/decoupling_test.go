// Package decoupling holds the §11.4.28 decoupling audit: the HelixStream
// engine (cmd/ + pkg/) MUST carry ZERO consumer-project literal — no ATMOSphere
// package prefixes, device serials, region-specific package ids, or
// ATMOSphere-tree import paths. All project-specific data is registered by the
// consumer at runtime via pkg/registry (RegisterAdapter / RegisterRoster /
// RegisterEndpoint), never baked into the engine.
//
// The forbidden needles are built from fragments so this audit file does not
// match itself; the audit also skips its own directory.
package decoupling

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// forbiddenNeedles are consumer-project literals that must never appear in the
// engine source. Built from fragments so this file itself does not self-match.
func forbiddenNeedles() []string {
	return []string{
		"com." + "atmosphere.",    // ATMOSphere Android package prefix
		"ru." + "kinopoisk",       // region-specific consumer package id
		"ru." + "start.androidtv", // region-specific consumer package id
		"998fd366" + "15e99484",   // consumer device serial (D3)
		"66ff9c4f" + "51f00ee7",   // consumer device serial (D4)
		"device/" + "rockchip",    // ATMOSphere-tree import/path
	}
}

// moduleRoot walks up from the test's working directory to the dir containing go.mod.
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
			t.Fatal("decoupling audit: could not locate module root (go.mod)")
		}
		dir = parent
	}
}

func TestEngineHasNoProjectLiteral(t *testing.T) {
	root := moduleRoot(t)
	needles := forbiddenNeedles()

	// Scan every .go file under cmd/ and pkg/ (the reusable engine). Skip this
	// audit's own directory (it deliberately names the needles).
	scanRoots := []string{filepath.Join(root, "cmd"), filepath.Join(root, "pkg")}
	var scanned int
	for _, sr := range scanRoots {
		err := filepath.Walk(sr, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || !strings.HasSuffix(path, ".go") {
				return nil
			}
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			scanned++
			content := string(b)
			for _, n := range needles {
				if strings.Contains(content, n) {
					t.Errorf("§11.4.28 decoupling violation: %s contains project literal %q", path, n)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", sr, err)
		}
	}
	if scanned == 0 {
		t.Fatal("decoupling audit scanned zero .go files — audit is not actually running")
	}
	t.Logf("§11.4.28 decoupling audit: scanned %d engine .go files, zero project literals", scanned)
}
