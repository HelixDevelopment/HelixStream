package evidence

import "testing"

// TestEvidenceValid is the anti-bluff table: a PASS is valid ONLY with a
// non-empty evidence Path; a SKIP/OPERATOR-BLOCKED is valid ONLY with a Reason.
func TestEvidenceValid(t *testing.T) {
	cases := []struct {
		name string
		ev   Evidence
		want bool
	}{
		{"pass with path", Evidence{Verdict: Pass, Path: "/qa/vid.json"}, true},
		{"pass without path is a bluff", Evidence{Verdict: Pass, Path: ""}, false},
		{"skip with reason", Evidence{Verdict: Skip, Reason: "topology_unsupported"}, true},
		{"skip without reason", Evidence{Verdict: Skip, Reason: ""}, false},
		{"operator-blocked with reason", Evidence{Verdict: OperatorBlocked, Reason: "otp timeout"}, true},
		{"operator-blocked without reason", Evidence{Verdict: OperatorBlocked}, false},
		{"fail is always valid", Evidence{Verdict: Fail}, true},
		{"unknown verdict is invalid", Evidence{Verdict: "MAYBE"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.ev.Valid(); got != c.want {
				t.Fatalf("Valid()=%v want %v for %+v", got, c.want, c.ev)
			}
		})
	}
}

func TestConstructors(t *testing.T) {
	if p := NewPass(VideoDisplay, "/qa/x", "d"); !p.Valid() || p.Verdict != Pass || p.Feature != VideoDisplay {
		t.Fatalf("NewPass produced invalid record: %+v", p)
	}
	if s := NewSkip(Geo, "geo_restricted", "d"); !s.Valid() || s.Verdict != Skip {
		t.Fatalf("NewSkip produced invalid record: %+v", s)
	}
	// A PASS built without a path must be caught by Valid — the constructor does
	// not fabricate one.
	if bad := NewPass(AudioOutput, "", "d"); bad.Valid() {
		t.Fatal("NewPass with empty path must NOT be Valid (anti-bluff)")
	}
	if f := NewFail(SubtitleRender, "d"); !f.Valid() || f.Verdict != Fail {
		t.Fatalf("NewFail produced invalid record: %+v", f)
	}
}
