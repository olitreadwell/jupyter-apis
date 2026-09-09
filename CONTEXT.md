# StatCan/jupyter-apis context
> refreshed 2026-09-09 | upstream default: main @ 694a5a2 | upstream active dev: jupyter-apis-aaw2.0 @ b5451b2

## Identity & policies
- upstream: StatCan/jupyter-apis, default branch `main` (stale, last touched 2025-12-30), active dev branch `jupyter-apis-aaw2.0` (all PRs since at least #360 base here). Primary languages: Go backend + Angular 17 (TypeScript) frontend. English-first (yes; README/CONTRIBUTING bilingual EN/FR, issues in English).
- CLA/DCO: none. contributing: CONTRIBUTING.md present (asks for issue discussion, no signup).
- AI-assisted PR policy: unstated (no ban, no disclosure requirement found).
- signed commits required: no.
- PR template: none (passport pr_template_present false); use pipeline body fallback.
- external tracker: GitHub issues, but maintainers reference internal Jira (jirab.statcan.ca, e.g. issue #400 -> ZONE-506) inside issue/PR bodies. Not a hard tracker gate.

## Conventions (verified from merged PRs)
- branch naming: upstream merged-PR headRefNames are plain kebab (`volume-pending-resize`, `expand-pvc-frontend`, `feat-oom-msg`, `backend-pvc-increase`, `update-default-notebook-icon`); feature/area prefixes common. Fall back to `type/desc`.
- commit style: conventional-ish (`feat(frontend): ...`, `feat(edit-notebook): ...`, `bug(UI): ...`, `chore(go): ...`).
- base for EVERY upstream PR is `jupyter-apis-aaw2.0`, NOT `main`. A fork PR targeting `main` cannot be promoted cleanly and its Build CI does not trigger (build.yml only fires on PRs into aaw2.0). Target `jupyter-apis-aaw2.0`.
- CI: Build (Go `go build` + `go vet` via task, only on PRs into aaw2.0), JWA Frontend Tests + Common Frontend Tests (prettier format:check, ng lint, karma unit tests, cypress). Go backend currently has ZERO unit tests.
- responsiveness: high — numerous external and team PRs merged Aug 2026, typically within days.

## Maintainer picture
- active maintainers: mathis-marcotte (very active, most recent work), wg102, Jose-Matsuda. Good outside-merge evidence (wg102 etc. merged).

## Issue-area health
- volumes / forms / size options = IN FLUX + claimed: open PR #400 (mathis-marcotte) "changed the data type of the size options in the form" (auto-deploy). Do NOT touch form-cpu-ram / volume size controls.
- issue #347 (SAS defaults below validator minimums): DROPPED-ALREADY-FIXED by commit 911bc0ea (2025-02-18) which lowered validators to match deployed config; stale screenshot in issue. No work remains.
- No maintainer-engaged (documented + approved + open) issue survives. Use repo-audit self-found gaps only.

## Gap ledger (dedupe — READ FIRST, never re-pick)
- 2026-08-05 PR #1 (fork) test-coverage `configSizeToNumber`/`calculateLimits` — pr-opened-substantive; Go Build red = k3d v4.4.7 404 (env/toolchain, unrelated).
- 2026-08-12 PR #2 (fork) test-coverage form-new size utils (base `main`) — open; body contains an AI-assistance disclosure (now out of policy) and targets stale `main`. Needs hygiene (strip AI mention, retarget aaw2.0) in a future pass.
- 2026-08-24 issue #347 — dropped-already-fixed (commit 911bc0ea). Lesson: always verify live, never trust the issue's stale screenshot.
- 2026-09-09 README accuracy fix — pr-opened (fork, branch off jupyter-apis-aaw2.0). Distinct from PR #1/#2.
- 2026-09-09 backend validation errors (notebooks.go) — pr-opened https://github.com/olitreadwell/jupyter-apis/pull/4 (fix: preserve all notebook validation errors; branch fix-backend-validation-errors, base jupyter-apis-aaw2.0). Verified live in upstream b5451b2: validateNotebook/validateUpdateNotebook used assignment instead of append for resource + data-volume errors, silently dropping earlier errors (missing name/namespace/image, invalid name, zero cpu). Added notebooks_test.go. CI green (Unit tests, Check code format and lint, Test, UI tests with Cypress all success).

## Mined gaps (discovered, not yet attempted)
- 2026-09-09 docs README stale/wrong references: `thunder-tests` -> `rest-tests` folder, `THUNDER CLIENT` -> `REST Client` extension, `rest-test\restclient.http` -> `rest-tests\restclient.http` (x2), `intergration`->integration heading, `recommanded`->recommended, `email adress`/`dev cluser` typos, dead link `./DELTA_FRONTEND.md` -> real file `./DEALTA_FRONTEND.md`. All verified present in upstream b5451b2. — status: attempted / PR OPENED https://github.com/olitreadwell/jupyter-apis/pull/3 (docs: fix stale and misspelled README references)
