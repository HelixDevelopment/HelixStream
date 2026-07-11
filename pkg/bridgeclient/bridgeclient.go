// Package bridgeclient defines the Claude-Code ↔ HelixQA bridge +
// operator-interaction contract (DESIGN.md §9). It is the client side of the
// loopback bridge the engine uses to (a) ask the driving agent a nav/UX
// question the vision layer cannot resolve, (b) request an interactive
// SMS/email OTP or PIN from the operator, and (c) route a UI/UX reasoning
// prompt to the model the Claude Code session exposes.
//
// Credentials discipline (§11.4.10): an OTP transits the loopback bridge ONLY
// and is NEVER logged, printed, or committed. RequestOTP returns the code to the
// caller in-memory; callers MUST NOT log it. A masked destination (e.g.
// "+7•••••34") is the only destination info that may appear anywhere.
//
// P1 NOTE (skeleton): the contract + a stub that models the request/response +
// the bounded-timeout → OPERATOR-BLOCKED path (§11.4.21). The real transport
// (extending helixqa-bridge with /v1/interaction/* + /v1/model/complete) lands
// in P6.
package bridgeclient

import (
	"context"
	"errors"
)

// OTPRequest asks the operator for an interactive code. It carries ONLY a masked
// destination — never the code, never the full number.
type OTPRequest struct {
	Service           string
	Delivery          string // "sms" | "email"
	Hint              string // e.g. "6-digit"
	DestinationMasked string // e.g. "+7•••••34"
}

// NavQuestion asks the driving agent to choose among candidate actions when the
// vision layer cannot resolve navigation (§11.4.117 blank-hierarchy fallback).
type NavQuestion struct {
	FrameRef   string   // reference to a captured frame (path/id), not raw pixels
	Candidates []string // candidate actions the agent chooses among
}

var (
	// ErrOperatorBlocked is returned when a bounded interaction times out with no
	// operator response — the engine records OPERATOR-BLOCKED (§11.4.21), never a guess.
	ErrOperatorBlocked = errors.New("bridgeclient: operator interaction timed out (operator-blocked)")
	// ErrNoCandidates is returned by AskNav when a question has no candidates.
	ErrNoCandidates = errors.New("bridgeclient: nav question has no candidates")
)

// OperatorInteraction is the operator/model interaction half of the bridge.
type OperatorInteraction interface {
	// AskNav poses a navigation question and returns the chosen candidate.
	AskNav(ctx context.Context, q NavQuestion) (choice string, err error)
	// RequestOTP requests an interactive code/PIN. On a bounded timeout with no
	// answer it returns ErrOperatorBlocked. The returned code MUST NOT be logged.
	RequestOTP(ctx context.Context, req OTPRequest) (code string, err error)
}

// Bridge is the full bridge contract: operator interaction + model routing +
// health.
type Bridge interface {
	OperatorInteraction
	// ModelComplete routes a UI/UX reasoning prompt to the operator's model and
	// returns the next action (§11.4.160). Anti-bluff: the real bridge attaches
	// captured evidence to every response.
	ModelComplete(ctx context.Context, prompt string) (string, error)
	// Health checks the loopback bridge is up.
	Health(ctx context.Context) error
}

// StubBridge is a deterministic reference implementation for contract tests. It
// performs NO network I/O and stores NO secret.
type StubBridge struct {
	// OTPCode is the code the operator "supplies". Empty ⇒ RequestOTP returns
	// ErrOperatorBlocked (models the bounded-timeout with no answer).
	OTPCode string
	// ModelAction is what ModelComplete returns.
	ModelAction string
	// Healthy toggles Health().
	Healthy bool
}

func (b StubBridge) AskNav(_ context.Context, q NavQuestion) (string, error) {
	if len(q.Candidates) == 0 {
		return "", ErrNoCandidates
	}
	return q.Candidates[0], nil
}

func (b StubBridge) RequestOTP(_ context.Context, _ OTPRequest) (string, error) {
	if b.OTPCode == "" {
		return "", ErrOperatorBlocked
	}
	return b.OTPCode, nil
}

func (b StubBridge) ModelComplete(_ context.Context, _ string) (string, error) {
	if b.ModelAction == "" {
		return "", errors.New("bridgeclient: stub has no model action configured")
	}
	return b.ModelAction, nil
}

func (b StubBridge) Health(_ context.Context) error {
	if !b.Healthy {
		return errors.New("bridgeclient: bridge not healthy")
	}
	return nil
}

// Compile-time assertion that StubBridge satisfies Bridge.
var _ Bridge = StubBridge{}
