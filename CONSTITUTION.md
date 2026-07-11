# HelixStream Constitution

## INHERITED FROM the Helix Constitution

All rules in the Helix Constitution's `Constitution.md` (and the `CLAUDE.md`
/ `AGENTS.md` it references) apply unconditionally. The submodule-scoped
rules below extend the universal clauses — they MUST NOT weaken any
inherited rule. When consumed as a submodule, locate the constitution from
any nested depth via the parent project's `constitution/find_constitution.sh`
helper — do NOT hardcode a path (this module stays fully decoupled and
project-agnostic per §11.4.28).

Canonical reference: https://github.com/HelixDevelopment/HelixConstitution

## Anti-bluff — the binding mandate for this module

HelixStream is a universal, reusable audio/video streaming-app testing
System, part of HelixQA. Tests and Challenges exist for exactly one purpose:
to confirm a streaming feature genuinely works for a real end user,
end-to-end. A PASS on a broken feature is a bluff and is forbidden (§11.4).
Every PASS this System emits — and every PASS in its own suite — MUST carry
positive captured evidence per §11.4.5 / §11.4.69. Analyzers are
self-validated (§11.4.107(10)); gates are mutation-paired (§1.1);
credentials are never committed or logged (§11.4.10); a geo-blocked service
SKIPs-with-reason, never a fake PASS and never a FAIL-for-geo (§11.4.3).

## Decoupling (§11.4.28)

The engine carries no consumer-project literal; the consumer registers its
data at runtime via `pkg/registry`. The `internal/decoupling` audit enforces
this. This module is reusable by ANY project, not just its first consumer.

See [`CLAUDE.md`](CLAUDE.md) / [`AGENTS.md`](AGENTS.md) / [`QWEN.md`](QWEN.md)
/ [`GEMINI.md`](GEMINI.md) for the full operating manual, and
`docs/P1_STATUS.md` for the honest current build state.
