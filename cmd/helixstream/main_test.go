package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunVersion(t *testing.T) {
	var out, errb bytes.Buffer
	if code := run([]string{"version"}, &out, &errb); code != 0 {
		t.Fatalf("version exit = %d, want 0", code)
	}
	if !strings.Contains(out.String(), Version) {
		t.Fatalf("version output %q does not contain %q", out.String(), Version)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var out, errb bytes.Buffer
	if code := run([]string{"frobnicate"}, &out, &errb); code != 2 {
		t.Fatalf("unknown-command exit = %d, want 2", code)
	}
	if !strings.Contains(errb.String(), "unknown command") {
		t.Fatalf("stderr %q missing 'unknown command'", errb.String())
	}
}

func TestRunNoArgs(t *testing.T) {
	var out, errb bytes.Buffer
	if code := run(nil, &out, &errb); code != 2 {
		t.Fatalf("no-args exit = %d, want 2", code)
	}
}

// A stub subcommand must exit 0 AND explicitly announce it is not implemented —
// never silently pretend to work (anti-bluff). `validate` is NO LONGER a stub.
func TestRunStubSubcommandIsHonest(t *testing.T) {
	for _, cmd := range []string{"run", "adapters", "record", "login"} {
		var out, errb bytes.Buffer
		if code := run([]string{cmd}, &out, &errb); code != 0 {
			t.Fatalf("%s exit = %d, want 0", cmd, code)
		}
		if !strings.Contains(errb.String(), notImplemented) {
			t.Fatalf("%s must announce %q, got stderr=%q", cmd, notImplemented, errb.String())
		}
	}
}

func TestRunHelp(t *testing.T) {
	var out, errb bytes.Buffer
	if code := run([]string{"--help"}, &out, &errb); code != 0 {
		t.Fatalf("--help exit = %d, want 0", code)
	}
	if !strings.Contains(out.String(), "usage:") {
		t.Fatalf("--help output missing usage: %q", out.String())
	}
}
