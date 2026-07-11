# HelixStream

**Universal, reusable audio/video streaming-app testing System for [HelixQA](https://github.com/HelixDevelopment/HelixQA).**

> Status: **P1 landed — skeleton + wiring.** The reusable core interfaces (adapter contract, login state machine, catalog, playback, topology runner, geo-gate, smooth-AV acceptance, bridge/operator-interaction) are defined as Go types with deterministic stub implementations and green contract tests (`go build ./...` + `go test ./...` pass across 12 packages). **No real device driving is implemented** — the login/catalog/playback/smooth-AV/recording/bridge implementations are later PWUs. The authoritative design + 11-PWU plan lives (for now) in the consuming project at `docs/research/helixqa_streaming_test_20260711/DESIGN.md`; the honest per-PWU build state lives at [`docs/P1_STATUS.md`](docs/P1_STATUS.md). Nothing here validates a real streaming app yet — claiming otherwise would be the exact §11.4 PASS-bluff this System exists to prevent.

## What it is

A project-agnostic (Constitution §11.4.28) library + knowledge base that gives **any** project out-of-the-box means to autonomously test audio/video streaming services and apps on Android / Android TV:

- **Per-service adapters** (login state machine, catalog map, playback-control map), **version-tracked** to survive UI drift.
- **Login / multi-step-auth** including an **interactive SMS/email-code operator handoff** (the code is requested from the operator through the driving agent and fed back).
- **Catalog browse** + **playback manipulation**: choose highest audio+video tracks, rewind to random positions, pause/resume/stop, select audio/video/subtitle tracks — with **human-like** pacing.
- **Topology-aware dual-display validation** (§11.4.3): content on the 2nd display by default + subtitles, UI-only on the 1st, primary switch-button toggles both between displays — runs only on dual-display devices, `topology_unsupported` SKIP on single-display.
- **Live recording of all connected displays** + **anti-bluff real-subtitle-content validation** (real dialogue, never placeholders like "You're watching video").
- A **Claude-Code ↔ HelixQA bridge** so the driving agent supplies the models HelixQA needs for UI/UX navigation and interactive code prompts.

## Composition, not reinvention (§11.4.74)

HelixStream **orchestrates** existing HelixQA capabilities (`helixqa-bridge`, `recording-analyzer`, `helixqa-omniparser`/`helixqa-uitars` vision, `helixqa-recvalidate`, `helixqa-input`, `qa-audio-probe`, `pkg/autonomous`) and the host dual-display recording pipeline. It does not duplicate them.

## License

Apache-2.0.
