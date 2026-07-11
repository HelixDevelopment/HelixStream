# HelixStream — P1 status (skeleton + wiring)

**Revision:** 1
**Last modified:** 2026-07-11
**Scope:** DESIGN.md §12 PWU **P1 (Module skeleton + wiring)** — ONLY.
**Track/label:** (T1/main - claude4)

## What P1 landed (real, compiles + tests green)

P1 is the **skeleton + wiring** milestone. It defines the reusable core
interfaces as Go types with deterministic stub implementations and contract
tests, and wires the module as an own-org dependency. It implements **no**
real device driving — that is the honest boundary.

### Core interfaces defined (the NEW reusable layer, DESIGN.md §3.2)

| Package | Core interface | Role |
|---|---|---|
| `pkg/adapter` | `ServiceAdapter` (composes the three below) | per-service adapter contract (DESIGN.md §10) |
| `pkg/loginsm` | `LoginStateMachine` | login FSM incl. `INTERACTIVE_CODE_HANDOFF` + PIN (DESIGN.md §6) |
| `pkg/catalog` | `CatalogMap` | catalog browse (DESIGN.md §7) |
| `pkg/playback` | `PlaybackController` | highest-track / seek / transport / track select (DESIGN.md §7) |
| `pkg/topology` | `Detector` + `Runner` | topology-parameterized runner (§11.4.3, DESIGN.md §8) |
| `pkg/geogate` | `GeoGate` | reachability-first geo/VPN gate (DESIGN.md §8.2) |
| `pkg/smoothav` | `Acceptance` + `Analyzer` | smooth-AV oracle, HW ALSA PCM advancing (DESIGN.md §8.1) |
| `pkg/bridgeclient` | `OperatorInteraction` + `Bridge` | Claude-Code ↔ HelixQA bridge + OTP handoff (DESIGN.md §9) |
| `pkg/registry` | `Registry` | **the §11.4.28 decoupling seam** — consumer registers adapters/roster/endpoints at runtime |
| `pkg/evidence` | `Evidence` + `Valid()` | §11.4.69 captured-evidence shapes + anti-bluff invariant |

`cmd/helixstream` — CLI; only `version` is implemented, the other
subcommands print an explicit `NOT IMPLEMENTED (P1 skeleton)` marker.

### Anti-bluff proofs captured at P1

- `go build ./...` = 0, `go vet ./...` = 0, `go test -count=1 ./...` = 0 across
  **12 packages / 65 test cases** (stdlib-only, no network).
- `internal/decoupling` audit (§11.4.28): scanned **22 engine .go files**, **zero
  consumer-project literals** (no `com.atmosphere.*`, device serials, region
  package ids, or `device/rockchip` paths).
- `pkg/smoothav` **self-validation** (§11.4.107(10)): the acceptance analyzer
  PASSes its golden-good fixtures and FAILs its golden-bad fixtures — including
  the **AudioFlinger-software-counter bluff** case (HW `hw_ptr` NOT advancing →
  FAIL even when the SW "written" counter looks fine).
- `pkg/evidence` anti-bluff invariant: a PASS with no evidence path is `Valid()==false`.

## What P1 did NOT do (honest boundary, §11.4.6)

Every package is **skeleton + stub**. There is **no** real device interaction,
no recording, no vision nav, no ADB, no Mullvad, no bridge transport, no
adapter for any real service. Nothing here validates a real streaming app.

## Next steps (later PWUs, DESIGN.md §12)

- **P2** — bank schema + loader + generic `banks/streaming_*.yaml`.
- **P3** — `pkg/topology`/`pkg/displayval` real detection + `pkg/record` glue to the host dual-display recorder.
- **P3b** — `pkg/geogate` real on-device geo probe + Mullvad backend (device-kernel WireGuard UNCONFIRMED — verify `/proc/config.gz`).
- **P3c** — `pkg/smoothav` real metric sources (recording-analyzer + qa-audio-probe + `/proc/asound` reader).
- **P4** — `pkg/subtitleoracle` port of `subtitle_content_validation.sh` (§11.4.137).
- **P5/P6** — login FSM driver + interactive-auth handoff over the extended `helixqa-bridge` (`/v1/interaction/*`, `/v1/model/complete`).
- **P7** — `pkg/catalog`/`pkg/playback`/`pkg/humanize` real drivers.
- **P8/P9** — video + music service adapters (consumer data).
- **P10** — constitution anchor §11.4.193 propagation.
- **P11** — ATMOSphere consumer checkset + full roster migration.

## Dependency wiring (§11.4.31)

`helix-deps.yaml` declares the intended own-org deps (HelixQA, Containers,
VisionEngine). The P1 skeleton compiles **standalone** (stdlib-only `go.mod`);
the `require`/`replace` block is added when the glue packages that import
those deps land (P3/P6/P7). Consumed into the ATMOSphere tree as a submodule
at `tools/helixqa/helix_stream` (alongside its HelixQA sibling deps).
