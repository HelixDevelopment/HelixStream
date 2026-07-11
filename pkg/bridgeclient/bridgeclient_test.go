package bridgeclient

import (
	"context"
	"testing"
)

func TestAskNav(t *testing.T) {
	b := StubBridge{}
	ctx := context.Background()

	got, err := b.AskNav(ctx, NavQuestion{FrameRef: "f1", Candidates: []string{"play", "back"}})
	if err != nil || got != "play" {
		t.Fatalf("AskNav = %q err=%v, want play", got, err)
	}
	if _, err := b.AskNav(ctx, NavQuestion{FrameRef: "f1"}); err != ErrNoCandidates {
		t.Fatalf("AskNav no candidates = %v, want ErrNoCandidates", err)
	}
}

func TestRequestOTP(t *testing.T) {
	ctx := context.Background()
	req := OTPRequest{Service: "svc", Delivery: "sms", Hint: "6-digit", DestinationMasked: "+7•••••34"}

	// Operator supplies a code.
	got, err := StubBridge{OTPCode: "123456"}.RequestOTP(ctx, req)
	if err != nil || got != "123456" {
		t.Fatalf("RequestOTP = %q err=%v, want 123456", got, err)
	}
	// No answer within the bounded window → OPERATOR-BLOCKED, never a guess.
	if _, err := (StubBridge{}).RequestOTP(ctx, req); err != ErrOperatorBlocked {
		t.Fatalf("RequestOTP timeout = %v, want ErrOperatorBlocked", err)
	}
}

func TestModelCompleteAndHealth(t *testing.T) {
	ctx := context.Background()
	if _, err := (StubBridge{ModelAction: "tap(play)"}).ModelComplete(ctx, "what next?"); err != nil {
		t.Fatalf("ModelComplete: %v", err)
	}
	if _, err := (StubBridge{}).ModelComplete(ctx, "x"); err == nil {
		t.Fatal("ModelComplete with no configured action must error")
	}
	if err := (StubBridge{Healthy: true}).Health(ctx); err != nil {
		t.Fatalf("Health(healthy) = %v", err)
	}
	if err := (StubBridge{Healthy: false}).Health(ctx); err == nil {
		t.Fatal("Health(unhealthy) must error")
	}
}
