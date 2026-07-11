// Package evidence defines the captured-evidence shapes and the verdict
// vocabulary shared across HelixStream.
//
// It encodes the anti-bluff discipline of the Helix Constitution: every PASS
// for a user-visible feature MUST cite a non-empty captured-evidence path
// (§11.4.5 / §11.4.69 sink-side taxonomy), and every SKIP / OPERATOR-BLOCKED
// MUST carry a reason (§11.4.3). A PASS with no evidence path is a §11.4
// PASS-bluff; Evidence.Valid mechanises that rule so callers cannot emit one.
//
// P1 NOTE (skeleton): this package defines the types + the Valid invariant.
// The actual capture pipelines (recording-analyzer glue, qa-audio-probe glue,
// subtitle oracle) are wired in later PWUs (P3/P3c/P4). Nothing here captures
// anything on its own.
package evidence

// Verdict is the closed-set outcome vocabulary (§11.4.45 / §11.4.116).
type Verdict string

const (
	Pass            Verdict = "PASS"
	Fail            Verdict = "FAIL"
	Skip            Verdict = "SKIP"
	OperatorBlocked Verdict = "OPERATOR-BLOCKED"
)

// FeatureClass is a member of the §11.4.69 closed-set sink-side evidence
// taxonomy relevant to streaming validation. Additional classes exist in the
// constitution; these are the ones HelixStream asserts on.
type FeatureClass string

const (
	VideoDisplay    FeatureClass = "video_display"
	AudioOutput     FeatureClass = "audio_output"
	SubtitleRender  FeatureClass = "subtitle_render"
	DisplayTopology FeatureClass = "display_topology"
	Geo             FeatureClass = "geo"
)

// Evidence is a single captured-evidence record produced by an assertion.
//
// Anti-bluff invariant (Valid): a PASS MUST cite a non-empty Path; a SKIP or
// OPERATOR-BLOCKED MUST cite a non-empty Reason. A FAIL is always structurally
// valid (a genuine defect need not cite an artefact, though it usually will).
type Evidence struct {
	Feature FeatureClass
	Verdict Verdict
	Path    string // filesystem path to the captured artefact (mandatory for PASS)
	Reason  string // SKIP-with-reason / OPERATOR-BLOCKED reason (mandatory for those)
	Desc    string // human-readable description of what was asserted
}

// Valid reports whether the record satisfies the anti-bluff invariant.
func (e Evidence) Valid() bool {
	switch e.Verdict {
	case Pass:
		return e.Path != ""
	case Skip, OperatorBlocked:
		return e.Reason != ""
	case Fail:
		return true
	default:
		return false
	}
}

// NewPass builds a PASS record; the caller MUST supply a non-empty path or the
// record will fail Valid().
func NewPass(f FeatureClass, path, desc string) Evidence {
	return Evidence{Feature: f, Verdict: Pass, Path: path, Desc: desc}
}

// NewSkip builds a SKIP-with-reason record (§11.4.3).
func NewSkip(f FeatureClass, reason, desc string) Evidence {
	return Evidence{Feature: f, Verdict: Skip, Reason: reason, Desc: desc}
}

// NewFail builds a FAIL record.
func NewFail(f FeatureClass, desc string) Evidence {
	return Evidence{Feature: f, Verdict: Fail, Desc: desc}
}
