# v1.2 Milestone Plan

12 open issues. This document describes the recommended order of work and the reasoning behind it.

---

## Dependency Map

```
#619  Wails upgrade          ──── independent
#513  Circle of Corr. bug   ──── independent (tests plotly v3 upgrade too)
#480  Audit Phase 4 (tests) ──── completes #443
#443  Reliability audit      ──── closes when #480 is done
#484  Centralize config      ──── independent, touches same backend as #459
#459  PCAResult refactor     ──── must come AFTER #480 (test before refactor)
                                  affects: Go backends, JSON schemas, TS types, CLI
#503  App.tsx refactor       ──── must come AFTER #459 (types change)
#505  State management       ──── depends on #503
#507  Logging solution        ──── can be bundled with #503/#505
#506  React performance      ──── depends on #505, informed by #617
#617  React 19 research      ──── ongoing/tracking, informs #506
#616  WASM epic              ──── architecture planning only for v1.2
```

---

## Recommended Order

### Stage 1 — Quick Wins & Foundation (do first)

#### ~~1. #619 — Upgrade Wails to v2.12.0~~ DONE (PR #623, merged to develop)
**Why first:** Security improvements (origin verification, URL validation). Low-risk, contained,
no frontend changes. Sets a clean dependency baseline before larger refactors begin.

#### 2. #513 — Circle of Correlations palette bug
**Why second:** Isolated, visible bug. Also serves as a smoke test that the plotly v3 upgrade
(#618, already merged) didn't introduce regressions in plot rendering. The root cause is
react-plotly.js's handling of many traces on re-render — worth understanding before the
larger frontend refactors.

---

### Stage 2 — Backend Correctness (before any refactoring)

#### 3. #480 — Transform/preprocessing validation tests (Phase 4 of #443)
**Why before refactoring:** Never refactor code whose behavior you haven't locked down with
tests first. Phase 4 tests validate the transform operation and preprocessing pipeline — the
very code that #459 (PCAResult refactor) will restructure.

Key areas to test:
- Transform on new data produces correct scores
- Preprocessing is preserved through fit/transform cycle
- Inverse transform reconstruction accuracy
- Missing value handling in transform

Once #480 is merged, #443 (the reliability audit umbrella) can be closed.

#### 4. #484 — Centralize hard-coded parameters
**Why here:** Touches the same backend files as #459 but is a smaller, lower-risk change.
Better to move the hard-coded values (NIPALS tolerance, max iterations, eigenvalue thresholds)
into config before the structural refactor in #459 so they don't get lost or re-scattered.

Config targets:
- `internal/config/config.go` — algorithm parameters (NIPALS tolerance=1e-8, maxIter=1000)
- `internal/config/gui_config.go` — GUI thresholds
- `internal/core/kernel_pca.go` — eigenvalue threshold (1e-10)

---

### Stage 3 — Backend Refactoring (the biggest structural change)

#### 5. #459 — Refactor PCAResult Structure
**Why here, not earlier:** The reliability tests (#480) must be green first — they anchor
current behavior. If we refactor first and tests break, we can't tell if it's a bug or an
expected change.

**This is the highest-impact issue in the milestone.** The current struct has duplicate
and inconsistent fields (`ExplainedVar` vs `ExplainedVariance`, percentages vs fractions,
etc.) that violate the mathematical correctness principle. A clean struct here pays dividends
in every other issue.

Blast radius:
- `pkg/types/pca.go` — struct definition
- All four PCA backends (`pca.go`, `kernel_pca.go`, `temporal_pca.go`, NIPALS)
- `schemas/v1/*.json` — JSON schema must be updated
- `pkg/validation/` — schema validator
- `cmd/gopca-desktop/app.go` — Wails bindings
- Frontend TypeScript types (all visualization components that consume PCAResult)

Plan in sub-steps within the issue's PR:
1. Define new struct (keep old fields as deprecated aliases temporarily)
2. Update all backends to populate new fields
3. Update JSON schemas
4. Update frontend TS types
5. Remove deprecated aliases
6. Run full sklearn validation suite to confirm mathematical correctness preserved

---

### Stage 4 — Frontend Refactoring (after types are stable)

Do these in order — each creates the prerequisite for the next.

#### 6. #503 — App.tsx refactor (<500 lines target)
**Why after #459:** If PCAResult types change (Stage 3), any refactored code consuming
those types would need updating again. Do the type changes first, then refactor cleanly
against stable types.

Current state: 1783 lines. Approach from the issue:
- Phase 1: Extract utilities and helpers to `utils/`
- Phase 2: Create custom hooks (`usePCAConfig`, `useFileData`, `useGoCSVIntegration`, etc.)
- Phase 3: Extract JSX sections into sub-components

#### 7. #507 — Replace console statements with logging solution
**Bundle with #503:** Since App.tsx is being opened and restructured, replace the 13
`console.*` calls in the same PR rather than opening the file again later. The logging
solution (structured log levels, production-safe) can be a tiny shared utility.

#### 8. #505 — Implement proper state management
**After #503:** Much easier to introduce context/reducers once the component is already
split into hooks. The 27 `useState` hooks naturally group into the context slices described
in the issue (PCA config, file data, UI state, etc.).

#### 9. #506 — Optimize React performance
**After #505:** Proper state management eliminates a large class of unnecessary re-renders
on its own. Profile *after* #505 is merged before adding manual `useMemo`/`useCallback` —
some of the identified bottlenecks may already be resolved.

Also check #617 conclusions before proceeding: if React 19 with the Compiler is now
realistic, some manual memoization may be premature.

---

### Stage 5 — Research & Future (ongoing)

#### 10. #617 — React 19 Upgrade Research
The issue's own recommendation is "wait 3-6 months" from November 2025 — which puts
reassessment around May 2026. The key blocker is `react-plotly.js` compatibility
(unmaintained). By the time v1.2 frontend work is done, reassess:
- Has `react-plotly.js` been updated?
- Has the community validated plotly + React 19?
- Is the React Compiler stable enough to replace manual memoization (#506)?

Keep as a tracking issue. Do not start upgrade work without revisiting the analysis.

#### 11. #616 — WebAssembly Web Application (Epic)
This is a major architectural initiative. For v1.2, limit scope to:
- Architecture decision record (ADR) in `docs/devel/`
- Proof-of-concept of Go PCA engine compiled to WASM
- Assess bundle size and performance characteristics

Full implementation is v1.3+ work.

---

## Summary Table

| # | Issue | Stage | Effort | Depends on |
|---|-------|-------|--------|-----------|
| ~~#619~~ | ~~Wails v2.12.0 upgrade~~ DONE | 1 | Small | — |
| #513 | Circle of Corr. palette bug | 1 | Small | — |
| #480 | Transform/preprocessing tests | 2 | Medium | — |
| #443 | Reliability audit (close) | 2 | — | #480 |
| #484 | Centralize hard-coded params | 2 | Small | — |
| #459 | PCAResult refactor | 3 | Large | #480 |
| #503 | App.tsx refactor | 4 | Large | #459 |
| #507 | Logging solution | 4 | Small | #503 (bundle) |
| #505 | State management | 4 | Medium | #503 |
| #506 | React performance | 4 | Medium | #505 |
| #617 | React 19 research | 5 | — | ongoing |
| #616 | WASM epic | 5 | Large | planning only |

---

## Key Risks

**#459 (PCAResult refactor)** is the most dangerous issue — it changes the contract between
backend and frontend. If done without adequate tests (#480) in place first, regressions will
be hard to detect. The JSON schema version may need bumping if the field names change in a
breaking way, which would affect model file compatibility.

**#616 (WASM)** scope must be actively managed. Without a clear boundary, it can consume
unlimited time. Keep it to ADR + PoC for v1.2.

**#617 (React 19)** should not be started without re-running the compatibility analysis
against current versions of `react-plotly.js` and plotly v3.5 (now installed).

---

*Created: 2026-04-09*
