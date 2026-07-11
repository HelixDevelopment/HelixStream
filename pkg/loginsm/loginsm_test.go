package loginsm

import (
	"context"
	"testing"
)

// drive runs the FSM from START to a terminal state, feeding one Observation
// per step, and returns the terminal state + step count.
func drive(t *testing.T, obs Observation, maxSteps int) (State, int) {
	t.Helper()
	m := NewStub()
	ctx := context.Background()
	steps := 0
	for !m.Terminal() && steps < maxSteps {
		if _, err := m.Advance(ctx, obs); err != nil {
			t.Fatalf("Advance error at step %d: %v", steps, err)
		}
		steps++
	}
	return m.Current(), steps
}

func TestLoginPaths(t *testing.T) {
	cases := []struct {
		name string
		obs  Observation
		want State
	}{
		{"already authed short-circuits", Observation{AlreadyAuthed: true}, StateAuthed},
		{"password no-2fa no-profile", Observation{Method: MethodPassword}, StateAuthed},
		{"password → profile → no pin", Observation{Method: MethodPassword, ProfileRequired: true}, StateAuthed},
		{"password → profile → pin", Observation{Method: MethodPassword, ProfileRequired: true, PINRequired: true}, StateAuthed},
		{"code handoff, code received+accepted", Observation{Method: MethodCode, CodeReceived: true, CodeAccepted: true}, StateAuthed},
		{"code handoff, no code → auth failed (operator-blocked)", Observation{Method: MethodCode, CodeReceived: false}, StateAuthFailed},
		{"code handoff, code rejected → auth failed", Observation{Method: MethodCode, CodeReceived: true, CodeAccepted: false}, StateAuthFailed},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, steps := drive(t, c.obs, 20)
			if got != c.want {
				t.Fatalf("terminal state = %s, want %s (in %d steps)", got, c.want, steps)
			}
		})
	}
}

func TestInteractiveHandoffStateIsReached(t *testing.T) {
	// The load-bearing NEW state must be traversed on the code path.
	m := NewStub()
	ctx := context.Background()
	obs := Observation{Method: MethodCode, CodeReceived: true, CodeAccepted: true}
	seen := map[State]bool{}
	for !m.Terminal() {
		s, err := m.Advance(ctx, obs)
		if err != nil {
			t.Fatal(err)
		}
		seen[s] = true
	}
	if !seen[StateInteractiveCodeHandoff] {
		t.Fatal("code login path must traverse INTERACTIVE_CODE_HANDOFF")
	}
}

func TestAdvanceAfterTerminalErrors(t *testing.T) {
	m := NewStub()
	ctx := context.Background()
	obs := Observation{AlreadyAuthed: true}
	for !m.Terminal() {
		if _, err := m.Advance(ctx, obs); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := m.Advance(ctx, obs); err != ErrTerminal {
		t.Fatalf("expected ErrTerminal after terminal, got %v", err)
	}
}
