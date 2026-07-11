# HelixStream

**Universal, reusable audio/video streaming-app testing System for [HelixQA](https://github.com/HelixDevelopment/HelixQA).**

> Status: **SCAFFOLD / design-approved, implementation pending.** This repository was created as the foundation for a multi-phase program. The authoritative design + phased PWU plan lives (for now) in the consuming project at `docs/research/helixqa_streaming_test_20260711/DESIGN.md` and will be migrated here in Phase 1. Nothing here is implemented yet — see the design doc for the honest EXISTS-vs-NEW inventory.

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
