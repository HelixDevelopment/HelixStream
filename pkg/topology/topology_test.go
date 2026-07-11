package topology

import (
	"context"
	"testing"

	"digital.vasic.helixstream/pkg/evidence"
)

func TestDetector(t *testing.T) {
	for _, want := range []Topology{Single, Dual} {
		got, err := StubDetector{Topo: want}.Detect(context.Background())
		if err != nil || got != want {
			t.Fatalf("Detect = %v err=%v, want %v", got, err, want)
		}
	}
}

// The load-bearing §11.4.3 rule: a dual-only check SKIPs (never FAILs) on a
// single-display device.
func TestDispatchTopologyRule(t *testing.T) {
	ctx := context.Background()
	r := StubRunner{}

	single, err := r.Dispatch(ctx, Single)
	if err != nil {
		t.Fatal(err)
	}
	dual, err := r.Dispatch(ctx, Dual)
	if err != nil {
		t.Fatal(err)
	}

	// No FAIL may ever appear purely for want of a second display.
	for _, ev := range single {
		if ev.Verdict == evidence.Fail {
			t.Fatalf("single-display run must not FAIL: %+v", ev)
		}
		if !ev.Valid() {
			t.Fatalf("evidence not anti-bluff-valid: %+v", ev)
		}
	}

	// On single, the 2nd-display check must be a SKIP with the canonical reason.
	if got := find(single, evidence.DisplayTopology); got.Verdict != evidence.Skip || got.Reason != ReasonTopologyUnsupported {
		t.Fatalf("single 2nd-display check = %+v, want SKIP/%s", got, ReasonTopologyUnsupported)
	}
	// On dual, the same check runs (PASS with evidence).
	if got := find(dual, evidence.DisplayTopology); got.Verdict != evidence.Pass || got.Path == "" {
		t.Fatalf("dual 2nd-display check = %+v, want PASS with evidence path", got)
	}
}

func find(evs []evidence.Evidence, f evidence.FeatureClass) evidence.Evidence {
	for _, e := range evs {
		if e.Feature == f {
			return e
		}
	}
	return evidence.Evidence{}
}
