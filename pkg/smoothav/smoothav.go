// Package smoothav defines the smooth-AV acceptance contract (DESIGN.md §8.1):
// a streaming PASS requires SMOOTH audio+video, not merely "it played".
//
//   - VIDEO: measured FPS within tolerance of source, no dropped-frame burst,
//     no freeze (SSIM > 0.99 across a window = frozen ⇒ FAIL), jitter under budget.
//   - AUDIO: the HARDWARE ALSA PCM must be RUNNING with hw_ptr ADVANCING (read
//     from /proc/asound/.../status) — NOT AudioFlinger's software "written"
//     counter. Forensic anchor (D3 MPV): AudioFlinger reported 8-ch frames
//     "written" while the hardware PCM never opened — a sink-side bluff the
//     software counter hides. Plus per-channel RMS above floor + XRUN under budget.
//
// §11.4.107(10): the Analyzer is SELF-VALIDATED — it MUST PASS a golden-good
// fixture and FAIL a golden-bad fixture (incl. the AudioFlinger-bluff case). An
// analyzer that passes its golden-bad fixture is itself a bluff.
//
// P1 NOTE (skeleton): the contract + a threshold Analyzer + its self-validation.
// The real metric SOURCES (recording-analyzer, qa-audio-probe, /proc/asound
// reader) are wired in P3c.
package smoothav

import (
	"errors"
	"fmt"

	"digital.vasic.helixstream/pkg/evidence"
)

// VideoMetrics is the measured video sample fed to the acceptance check.
type VideoMetrics struct {
	MeasuredFPS   float64
	SourceFPS     float64
	DroppedFrames int     // decoder drop count over the window
	MinSSIM       float64 // min SSIM across consecutive frames in the window
	JitterMS      float64 // inter-frame presentation-time jitter (variance proxy)
}

// AudioMetrics is the measured audio sample fed to the acceptance check.
type AudioMetrics struct {
	PCMState       string  // ALSA /proc/asound state: "RUNNING" required
	HWPtrAdvancing bool    // hw_ptr increased across samples (the load-bearing signal)
	MinChannelRMS  float64 // min per-channel RMS (0..1); a dead channel is 0
	XRUNBurst      int     // XRUN/underrun count over the window
}

// Thresholds are the acceptance budgets. Real values are calibrated on the
// project's own fixtures (§11.4.6); these are conservative skeleton defaults.
type Thresholds struct {
	FPSTolerance float64 // allowed |measured-source|/source
	MaxDropped   int
	FreezeSSIM   float64 // SSIM >= this across the window ⇒ frozen
	MaxJitterMS  float64
	MinRMSFloor  float64
	MaxXRUNBurst int
}

// DefaultThresholds returns the skeleton default budgets.
func DefaultThresholds() Thresholds {
	return Thresholds{
		FPSTolerance: 0.10, // ±10 %
		MaxDropped:   3,
		FreezeSSIM:   0.99,
		MaxJitterMS:  8.0,
		MinRMSFloor:  0.01,
		MaxXRUNBurst: 2,
	}
}

// Acceptance is the core smooth-AV contract.
type Acceptance interface {
	// AssertVideoSmooth returns PASS only for genuinely smooth video.
	AssertVideoSmooth(vm VideoMetrics) evidence.Verdict
	// AssertAudioHWPCMRunning returns PASS only when the HARDWARE PCM is running
	// and advancing (never the AudioFlinger software counter).
	AssertAudioHWPCMRunning(am AudioMetrics) evidence.Verdict
}

// Analyzer is the threshold-based reference Acceptance implementation.
type Analyzer struct{ T Thresholds }

// NewAnalyzer builds an Analyzer with the default thresholds.
func NewAnalyzer() Analyzer { return Analyzer{T: DefaultThresholds()} }

func (a Analyzer) AssertVideoSmooth(vm VideoMetrics) evidence.Verdict {
	if vm.SourceFPS <= 0 {
		return evidence.Fail
	}
	if rel := abs(vm.MeasuredFPS-vm.SourceFPS) / vm.SourceFPS; rel > a.T.FPSTolerance {
		return evidence.Fail
	}
	if vm.DroppedFrames > a.T.MaxDropped {
		return evidence.Fail
	}
	if vm.MinSSIM >= a.T.FreezeSSIM { // too-similar consecutive frames = frozen
		return evidence.Fail
	}
	if vm.JitterMS > a.T.MaxJitterMS {
		return evidence.Fail
	}
	return evidence.Pass
}

func (a Analyzer) AssertAudioHWPCMRunning(am AudioMetrics) evidence.Verdict {
	// The load-bearing check: the HARDWARE PCM must be RUNNING and advancing.
	if am.PCMState != "RUNNING" || !am.HWPtrAdvancing {
		return evidence.Fail
	}
	if am.MinChannelRMS < a.T.MinRMSFloor { // a dead channel
		return evidence.Fail
	}
	if am.XRUNBurst > a.T.MaxXRUNBurst {
		return evidence.Fail
	}
	return evidence.Pass
}

// Golden fixtures for §11.4.107(10) self-validation.
func goldenGoodVideo() VideoMetrics {
	return VideoMetrics{MeasuredFPS: 30, SourceFPS: 30, DroppedFrames: 0, MinSSIM: 0.72, JitterMS: 2.0}
}
func goldenBadVideoFrozen() VideoMetrics {
	return VideoMetrics{MeasuredFPS: 30, SourceFPS: 30, DroppedFrames: 0, MinSSIM: 0.999, JitterMS: 2.0}
}
func goldenGoodAudio() AudioMetrics {
	return AudioMetrics{PCMState: "RUNNING", HWPtrAdvancing: true, MinChannelRMS: 0.3, XRUNBurst: 0}
}

// The AudioFlinger-bluff: SW counter "written" (looks fine) but the hardware PCM never advanced.
func goldenBadAudioAudioFlingerBluff() AudioMetrics {
	return AudioMetrics{PCMState: "OPEN", HWPtrAdvancing: false, MinChannelRMS: 0.3, XRUNBurst: 0}
}

// SelfValidate proves the Analyzer is not itself a bluff: it MUST PASS the
// golden-good fixtures and FAIL the golden-bad fixtures (§11.4.107(10)).
func (a Analyzer) SelfValidate() error {
	if v := a.AssertVideoSmooth(goldenGoodVideo()); v != evidence.Pass {
		return fmt.Errorf("smoothav self-validation: golden-good video not PASS (%s)", v)
	}
	if v := a.AssertVideoSmooth(goldenBadVideoFrozen()); v != evidence.Fail {
		return fmt.Errorf("smoothav self-validation: golden-bad frozen video not FAIL (%s)", v)
	}
	if v := a.AssertAudioHWPCMRunning(goldenGoodAudio()); v != evidence.Pass {
		return fmt.Errorf("smoothav self-validation: golden-good audio not PASS (%s)", v)
	}
	if v := a.AssertAudioHWPCMRunning(goldenBadAudioAudioFlingerBluff()); v != evidence.Fail {
		return errors.New("smoothav self-validation: golden-bad AudioFlinger-bluff audio not FAIL — the oracle is bluffing")
	}
	return nil
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}
