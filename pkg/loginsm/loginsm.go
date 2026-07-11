// Package loginsm defines the login state-machine contract for a streaming
// service adapter (DESIGN.md §6), including the interactive SMS/email-code +
// PIN operator-handoff states.
//
// The FSM is adapter-declared and engine-interpreted. The engine drives the
// transitions via input injection + vision navigation (later PWUs); this
// package defines ONLY the states, the observation input, the interface, and a
// deterministic stub that exercises the graph.
//
// P1 NOTE (skeleton): StubStateMachine encodes the transition graph so the
// contract is testable, but it drives NO real device — the real driver
// (helixqa-input + omniparser + the §8 bridge handoff) lands in P5/P6.
package loginsm

import (
	"context"
	"errors"
)

// State is one node of the login FSM (DESIGN.md §6 diagram).
type State string

const (
	StateStart                  State = "START"
	StateDetectAuth             State = "DETECT_AUTH"
	StateChooseMethod           State = "CHOOSE_METHOD"
	StateEnterCreds             State = "ENTER_CREDS"
	StateRequestCode            State = "REQUEST_CODE"
	StateAwait2FA               State = "AWAIT_2FA_CHALLENGE"
	StateInteractiveCodeHandoff State = "INTERACTIVE_CODE_HANDOFF" // ★ the NEW piece
	StateSubmitCode             State = "SUBMIT_CODE"
	StateSelectProfile          State = "SELECT_PROFILE"
	StateEnterPIN               State = "ENTER_PIN"
	StateAuthed                 State = "AUTHED"      // terminal (success)
	StateAuthFailed             State = "AUTH_FAILED" // terminal (failure)
)

// Method is the login method an adapter offers.
type Method string

const (
	MethodPassword Method = "password"
	MethodCode     Method = "code" // SMS / email OTP
)

// Observation is the context the engine feeds into one transition. Fields are
// the observed truth at the current step (never guessed — §11.4.6).
type Observation struct {
	AlreadyAuthed      bool
	Method             Method
	ChallengePresented bool // a 2FA challenge is on screen
	CodeReceived       bool // the operator supplied the OTP via the bridge handoff
	CodeAccepted       bool // the service accepted the submitted code
	ProfileRequired    bool
	PINRequired        bool
	PINAvailable       bool // PIN present in secrets file (else operator handoff)
}

// ErrTerminal is returned by Advance when the machine is already terminal.
var ErrTerminal = errors.New("loginsm: state machine is terminal")

// LoginStateMachine is the core login contract (one of the three sub-contracts
// that make up a ServiceAdapter).
type LoginStateMachine interface {
	// Current returns the current FSM state.
	Current() State
	// Advance performs exactly one transition given the observation and returns
	// the resulting state. It errors if already terminal.
	Advance(ctx context.Context, obs Observation) (State, error)
	// Terminal reports whether Current() is AUTHED or AUTH_FAILED.
	Terminal() bool
}

// StubStateMachine is a deterministic reference implementation of the FSM used
// for contract tests. It performs NO device interaction.
type StubStateMachine struct {
	state State
}

// NewStub returns a StubStateMachine positioned at START.
func NewStub() *StubStateMachine { return &StubStateMachine{state: StateStart} }

func (m *StubStateMachine) Current() State { return m.state }

func (m *StubStateMachine) Terminal() bool {
	return m.state == StateAuthed || m.state == StateAuthFailed
}

// Advance encodes the DESIGN.md §6 transition graph.
func (m *StubStateMachine) Advance(_ context.Context, obs Observation) (State, error) {
	if m.Terminal() {
		return m.state, ErrTerminal
	}
	m.state = next(m.state, obs)
	return m.state, nil
}

func next(s State, obs Observation) State {
	switch s {
	case StateStart:
		return StateDetectAuth
	case StateDetectAuth:
		if obs.AlreadyAuthed {
			return StateAuthed
		}
		return StateChooseMethod
	case StateChooseMethod:
		if obs.Method == MethodPassword {
			return StateEnterCreds
		}
		return StateRequestCode
	case StateEnterCreds:
		if obs.ChallengePresented {
			return StateAwait2FA
		}
		return afterCreds(obs)
	case StateRequestCode:
		return StateAwait2FA
	case StateAwait2FA:
		return StateInteractiveCodeHandoff
	case StateInteractiveCodeHandoff:
		if obs.CodeReceived {
			return StateSubmitCode
		}
		// bounded timeout with no code → operator-blocked terminal (§11.4.21):
		// modelled as AUTH_FAILED here; the engine records OPERATOR-BLOCKED.
		return StateAuthFailed
	case StateSubmitCode:
		if !obs.CodeAccepted {
			return StateAuthFailed
		}
		return afterCreds(obs)
	case StateSelectProfile:
		if obs.PINRequired {
			return StateEnterPIN
		}
		return StateAuthed
	case StateEnterPIN:
		return StateAuthed
	default:
		return s
	}
}

// afterCreds handles the post-authentication profile/PIN gates.
func afterCreds(obs Observation) State {
	if obs.ProfileRequired {
		return StateSelectProfile
	}
	if obs.PINRequired {
		return StateEnterPIN
	}
	return StateAuthed
}
