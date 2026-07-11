// Package topology defines the per-device display-topology detection + the
// topology-parameterized runner contract (DESIGN.md §8, Constitution §11.4.3).
//
// One bank must run correctly on BOTH single- and dual-display devices: on a
// dual-display device the 2nd-display validations run; on a single-display
// device those validations SKIP-with-reason `topology_unsupported` — they MUST
// NEVER FAIL for want of a second display (§11.4.3).
//
// P1 NOTE (skeleton): the Detector + Runner contracts + a stub. Real detection
// (`dumpsys display` + DRM connector status) and the real display validations
// land in P3.
package topology

import (
	"context"

	"digital.vasic.helixstream/pkg/evidence"
)

// Topology is the detected display configuration.
type Topology string

const (
	Single Topology = "single_display"
	Dual   Topology = "dual_display"
)

// ReasonTopologyUnsupported is the canonical §11.4.3 SKIP reason for a
// dual-display-only check on a single-display device.
const ReasonTopologyUnsupported = "topology_unsupported"

// Detector reports a device's current display topology.
type Detector interface {
	Detect(ctx context.Context) (Topology, error)
}

// Runner dispatches a topology-parameterized run and returns the per-check
// evidence. On Single it emits SKIP-with-reason topology_unsupported for
// dual-only checks (never FAIL); on Dual it runs them.
type Runner interface {
	Dispatch(ctx context.Context, topo Topology) ([]evidence.Evidence, error)
}

// StubDetector returns a fixed topology.
type StubDetector struct{ Topo Topology }

func (d StubDetector) Detect(context.Context) (Topology, error) { return d.Topo, nil }

// StubRunner models the topology dispatch for one dual-display-only check plus
// one always-runs check, so the §11.4.3 SKIP-not-FAIL rule is observable.
type StubRunner struct{}

func (StubRunner) Dispatch(_ context.Context, topo Topology) ([]evidence.Evidence, error) {
	// The single-display content check always runs.
	out := []evidence.Evidence{
		evidence.NewPass(evidence.VideoDisplay, "/qa/stub/primary_content.json", "primary display content liveness (stub)"),
	}
	// The 2nd-display + toggle check is dual-only.
	if topo == Dual {
		out = append(out,
			evidence.NewPass(evidence.DisplayTopology, "/qa/stub/second_display.json", "2nd-display content+subtitle+toggle (stub)"),
		)
	} else {
		out = append(out,
			evidence.NewSkip(evidence.DisplayTopology, ReasonTopologyUnsupported, "2nd-display check skipped on single-display device"),
		)
	}
	return out, nil
}
