package geogate

import (
	"context"
	"testing"
)

func TestEnsureReachabilityFirst(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name            string
		gate            StubGeoGate
		requiredCountry string
		wantOutcome     Outcome
		wantVPN         bool
		wantReason      string
	}{
		{
			name:            "native reachable → proceed, no VPN",
			gate:            StubGeoGate{NativeCountry: "KZ", ReachableNative: true, ReachableViaVPN: true},
			requiredCountry: "RU",
			wantOutcome:     ProceedNative,
			wantVPN:         false,
		},
		{
			name:            "blocked but VPN works → proceed via VPN",
			gate:            StubGeoGate{NativeCountry: "KZ", ReachableNative: false, ReachableViaVPN: true},
			requiredCountry: "US",
			wantOutcome:     ProceedVPN,
			wantVPN:         true,
		},
		{
			name:            "blocked, VPN in-scope but still unreachable → SKIP geo_restricted",
			gate:            StubGeoGate{NativeCountry: "KZ", ReachableNative: false, ReachableViaVPN: false},
			requiredCountry: "NO",
			wantOutcome:     SkipRestricted,
			wantReason:      ReasonGeoRestricted,
		},
		{
			name:            "blocked, no VPN path for country → SKIP unreachable_external",
			gate:            StubGeoGate{NativeCountry: "KZ", ReachableNative: false, ReachableViaVPN: false},
			requiredCountry: "any",
			wantOutcome:     SkipRestricted,
			wantReason:      ReasonUnreachableExt,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d, err := c.gate.Ensure(ctx, "https://svc.example/probe", c.requiredCountry)
			if err != nil {
				t.Fatal(err)
			}
			if d.Outcome != c.wantOutcome {
				t.Fatalf("outcome = %s, want %s", d.Outcome, c.wantOutcome)
			}
			if d.UsedVPN != c.wantVPN {
				t.Fatalf("usedVPN = %v, want %v", d.UsedVPN, c.wantVPN)
			}
			if c.wantReason != "" && d.Reason != c.wantReason {
				t.Fatalf("reason = %q, want %q", d.Reason, c.wantReason)
			}
			// A geo block must NEVER present as reachable-but-passed.
			if d.Outcome == SkipRestricted && d.Reachable {
				t.Fatal("SkipRestricted must not be Reachable")
			}
		})
	}
}

func TestDetectCountry(t *testing.T) {
	got, err := StubGeoGate{NativeCountry: "KZ"}.DetectCountry(context.Background())
	if err != nil || got != "KZ" {
		t.Fatalf("DetectCountry = %q err=%v", got, err)
	}
}
