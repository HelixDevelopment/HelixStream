# GEMINI.md

## INHERITED FROM the Helix Constitution

This module is governed by the Helix Constitution. All rules in the
constitution's `CLAUDE.md` and the `Constitution.md` it references apply
unconditionally — the universal anti-bluff covenant §11.4, the no-guessing
mandate §11.4.6, the credentials-handling mandate §11.4.10, host-session
safety §12, data safety §9, and mutation-paired gates §1.1. The
module-specific rules below extend the universal clauses; they never weaken
any of them. When this file disagrees with the constitution, the
constitution wins.

Canonical reference: https://github.com/HelixDevelopment/HelixConstitution

When consumed as a submodule, locate the constitution from any nested depth
via the parent project's `constitution/find_constitution.sh` helper — do NOT
hardcode a path (this module stays fully decoupled and project-agnostic per
§11.4.28).

---

# HelixStream — AI Agent Operating Manual

`digital.vasic.helixstream` is a **universal, reusable audio/video
streaming-app testing System**, part of HelixQA. It gives ANY project
out-of-the-box means to autonomously test streaming services + apps on
Android / Android TV: a per-service **adapter contract** (login state
machine + catalog map + playback controller), a **topology-parameterized
runner** (§11.4.3), a **geo/VPN gate** (reachability-first), a **smooth-AV
acceptance** oracle (hardware ALSA PCM advancing, never the AudioFlinger
software counter), and the **Claude-Code ↔ HelixQA bridge / operator
interaction** channel.

Authoritative design + 11-PWU plan: the consuming project's
`docs/research/helixqa_streaming_test_20260711/DESIGN.md`. Current build
state: `docs/P1_STATUS.md`.

## The binding mandate: anti-bluff (§11.4)

The bar is not "tests pass" but **"users can use the feature."** Every PASS
this System emits MUST carry positive captured evidence per §11.4.5 /
§11.4.69 (`video_display` / `audio_output` / `subtitle_render` /
`display_topology` / geo). A green summary line without that evidence is a
critical defect. Analyzers are self-validated with golden good/bad fixtures
(§11.4.107(10)) — an analyzer that passes its golden-bad fixture is itself a
bluff.

## Decoupling (§11.4.28) — the load-bearing invariant

The engine (`cmd/` + `pkg/`) carries **ZERO consumer-project literal** — no
package names, device serials, or region endpoints. The consumer registers
its adapters / roster / endpoints at runtime via `pkg/registry`
(`RegisterAdapter` / `RegisterRoster` / `RegisterEndpoint`). The
`internal/decoupling` audit test enforces this mechanically.

## Honest status boundary (§11.4.6)

P1 is **skeleton + wiring only** — the core interfaces + stubs compile and
their contract tests are green. The real login / catalog / playback /
smooth-AV / recording / bridge IMPLEMENTATIONS are later PWUs (see
`docs/P1_STATUS.md`). No claim of working streaming validation is made here;
claiming otherwise would be the exact §11.4 PASS-bluff this System exists to
prevent.
