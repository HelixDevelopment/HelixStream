package smoothav

import (
	"testing"

	"digital.vasic.helixstream/pkg/evidence"
)

func TestAssertVideoSmooth(t *testing.T) {
	a := NewAnalyzer()
	cases := []struct {
		name string
		vm   VideoMetrics
		want evidence.Verdict
	}{
		{"smooth", VideoMetrics{MeasuredFPS: 30, SourceFPS: 30, MinSSIM: 0.7, JitterMS: 2}, evidence.Pass},
		{"fps out of tolerance", VideoMetrics{MeasuredFPS: 20, SourceFPS: 30, MinSSIM: 0.7, JitterMS: 2}, evidence.Fail},
		{"dropped-frame burst", VideoMetrics{MeasuredFPS: 30, SourceFPS: 30, DroppedFrames: 10, MinSSIM: 0.7, JitterMS: 2}, evidence.Fail},
		{"frozen (ssim too high)", VideoMetrics{MeasuredFPS: 30, SourceFPS: 30, MinSSIM: 0.999, JitterMS: 2}, evidence.Fail},
		{"jitter over budget", VideoMetrics{MeasuredFPS: 30, SourceFPS: 30, MinSSIM: 0.7, JitterMS: 50}, evidence.Fail},
		{"zero source fps", VideoMetrics{MeasuredFPS: 30, SourceFPS: 0}, evidence.Fail},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := a.AssertVideoSmooth(c.vm); got != c.want {
				t.Fatalf("AssertVideoSmooth = %s, want %s", got, c.want)
			}
		})
	}
}

func TestAssertAudioHWPCMRunning(t *testing.T) {
	a := NewAnalyzer()
	cases := []struct {
		name string
		am   AudioMetrics
		want evidence.Verdict
	}{
		{"hw running + advancing", AudioMetrics{PCMState: "RUNNING", HWPtrAdvancing: true, MinChannelRMS: 0.3}, evidence.Pass},
		{"audioflinger bluff: not advancing", AudioMetrics{PCMState: "OPEN", HWPtrAdvancing: false, MinChannelRMS: 0.3}, evidence.Fail},
		{"running but hw_ptr not advancing", AudioMetrics{PCMState: "RUNNING", HWPtrAdvancing: false, MinChannelRMS: 0.3}, evidence.Fail},
		{"dead channel", AudioMetrics{PCMState: "RUNNING", HWPtrAdvancing: true, MinChannelRMS: 0.0}, evidence.Fail},
		{"xrun burst", AudioMetrics{PCMState: "RUNNING", HWPtrAdvancing: true, MinChannelRMS: 0.3, XRUNBurst: 9}, evidence.Fail},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := a.AssertAudioHWPCMRunning(c.am); got != c.want {
				t.Fatalf("AssertAudioHWPCMRunning = %s, want %s", got, c.want)
			}
		})
	}
}

// §11.4.107(10): the analyzer must pass its own golden good/bad self-validation.
func TestAnalyzerSelfValidation(t *testing.T) {
	if err := NewAnalyzer().SelfValidate(); err != nil {
		t.Fatalf("analyzer self-validation failed (the oracle itself is a bluff): %v", err)
	}
}
