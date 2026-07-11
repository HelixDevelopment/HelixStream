// Command helixstream is the HelixStream CLI (DESIGN.md §4).
//
// P1 NOTE (skeleton): only `version` is implemented. The `run | adapters |
// validate | record | login` subcommands are declared and print an explicit
// NOT-IMPLEMENTED marker so no one mistakes the skeleton for a working System.
// Each is landed in a later PWU (P2/P5/P7/P8). Reporting them as working would
// be the exact §11.4 PASS-bluff HelixStream exists to prevent.
package main

import (
	"fmt"
	"io"
	"os"
)

// Version is the HelixStream build version. P1 = the skeleton milestone.
const Version = "0.1.0-p1"

// notImplemented marks a declared-but-unimplemented P1 subcommand.
const notImplemented = "NOT IMPLEMENTED (P1 skeleton — see docs/P1_STATUS.md)"

// subcommands declared by the CLI. `version` is implemented; the rest are P1 stubs.
var subcommands = []string{"run", "adapters", "validate", "record", "login", "version"}

func usage(w io.Writer) {
	fmt.Fprintf(w, "helixstream %s — universal streaming-app testing System (skeleton)\n", Version)
	fmt.Fprintln(w, "usage: helixstream <command>")
	fmt.Fprintln(w, "commands:")
	fmt.Fprintln(w, "  version    print the HelixStream version (implemented)")
	fmt.Fprintln(w, "  run        run a streaming test bank            ("+notImplemented+")")
	fmt.Fprintln(w, "  adapters   list/inspect registered adapters     ("+notImplemented+")")
	fmt.Fprintln(w, "  validate   validate an adapter/bank spec         ("+notImplemented+")")
	fmt.Fprintln(w, "  record     record connected displays             ("+notImplemented+")")
	fmt.Fprintln(w, "  login      drive a service login state machine   ("+notImplemented+")")
}

// run dispatches the CLI and returns the process exit code, writing to the
// provided streams so it is unit-testable.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 {
		usage(stderr)
		return 2
	}
	switch args[0] {
	case "version":
		fmt.Fprintf(stdout, "helixstream %s\n", Version)
		return 0
	case "run", "adapters", "validate", "record", "login":
		fmt.Fprintf(stderr, "helixstream %s: %s\n", args[0], notImplemented)
		return 0
	case "-h", "--help", "help":
		usage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "helixstream: unknown command %q\n", args[0])
		usage(stderr)
		return 2
	}
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
