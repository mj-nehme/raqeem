---
name: Fix CI Lints & Typechecks
about: Diagnose and stabilize lint/typecheck jobs in CI
title: "[CI] Fix lints and typechecks"
labels: [ci, tooling, lint, typecheck]
assignees: mj-nehme
---

## Summary
Stabilize and fix CI jobs that run linters and typecheckers across backends and frontends. Ensure deterministic, cached, and fast runs.

## Context
- Workflow: `.github/workflows/ci.yml`
- These jobs should be green on PRs and `master`:
  - `lint-python` (ruff on Devices backend)
  - `typecheck-python` (mypy on Devices backend)
  - `lint-go` (golangci-lint on Mentor backend)
  - `lint-devices-frontend` (ESLint on Devices frontend)

## Affected Jobs (check all that apply)
- [ ] lint-python (ruff)
- [ ] typecheck-python (mypy)
- [ ] lint-go (golangci-lint)
- [ ] lint-devices-frontend (eslint)

## Reproduction
1. Open the latest failing run for the job above in GitHub Actions.
2. Paste the failing step output below (or attach as artifact excerpt).

```
<paste failing logs here>
```

## Expected vs Actual
- Expected: Job completes successfully with zero errors; warnings as configured.
- Actual: Describe the error(s) observed.

## Local Replication Commands

Python (Devices backend)
```bash
# Ruff
cd devices/backend/src
ruff check .

# mypy (relaxed config)
mypy app --ignore-missing-imports --no-color-output
```

Go (Mentor backend)
```bash
cd mentor/backend/src
go mod download
# If installed locally:
golangci-lint run --timeout=5m
# Or use docker image:
docker run --rm -v "$PWD":/app -w /app golangci/golangci-lint:v1.61.0 golangci-lint run --timeout=5m
```

JavaScript (Devices frontend)
```bash
cd devices/frontend
npm ci --prefer-offline --no-audit
npm run lint --silent
```

## Proposed Fix Checklist
- [ ] Identify and fix offending code (style/type issues) keeping repo conventions.
- [ ] Align tool versions with repo config to avoid drift:
  - Ruff/mypy versions from `pyproject.toml` / `requirements*.txt`.
  - Go: `actions/setup-go@v5` uses `1.25.x`; ensure local matches.
  - golangci-lint action uses `v2.6.1`; reproduce locally with a close version (e.g., v1.61.x container).
  - Node: `actions/setup-node@v4` uses Node 20; lock ESLint plugins accordingly.
- [ ] Ensure configs are respected:
  - Ruff: `[tool.ruff]` and `[tool.ruff.lint.*]` in `pyproject.toml`.
  - mypy: `[tool.mypy]` in `pyproject.toml`.
  - Go: `.golangci.yml` or project defaults if absent.
  - ESLint: `devices/frontend/eslint.config.js`.
- [ ] Improve CI determinism/perf (optional): verify caches present and scoped.
- [ ] Validate locally using the commands above.
- [ ] Validate in CI via a PR; ensure jobs pass on PR and `master`.

## Acceptance Criteria
- All selected jobs pass locally and in CI on multiple runs.
- No flakiness across at least two re-runs of the workflow.
- No loosening of lint/type rules unless justified and documented.
- CI caches working (pip/npm/golangci) without cache key conflicts.

## References
- CI workflow: `.github/workflows/ci.yml`
- Python config: `pyproject.toml` (`[tool.ruff]`, `[tool.mypy]`)
- Devices backend deps: `devices/backend/requirements.txt`, `requirements-test.txt`
- Mentor backend: `mentor/backend/src/go.mod`, `go.sum`
- Devices frontend: `devices/frontend/eslint.config.js`, `package.json`
