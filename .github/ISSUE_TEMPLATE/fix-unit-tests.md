---
name: Fix Unit Tests (Postgres only)
about: Make unit tests reliable without modifying the database; keep GORM models authoritative
labels: testing, backend, ci
assignees: mj-nehme
---

**Summary**
- Fix unit tests to run reliably with PostgreSQL.
- Do NOT modify the existing database schema or constraints during tests.
- Keep GORM models in `mentor/backend/src/models` as the single source of truth.

**Context**
- Project is learning/testing oriented; SQLite has been removed.
- Tests should always use Postgres locally and in CI.
- Recent failures include role/constraint errors and environment drift.

**Requirements**
- Always connect to PostgreSQL using environment variables: `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_DB`, `SSLMODE`.
- No migrations or DDL (ALTER/DROP) during unit test execution.
- Make tests deterministic and independent of the current DB state.
- Respect GORM models; do not change files under `mentor/backend/src/models`.

**Proposed Approach**
- Use transaction-based isolation for tests: begin a transaction in setup and rollback in teardown.
- Build DSN from env vars; parse port via `strconv.Atoi`; fail fast on missing critical vars.
- Remove ignoring of test failures in CI (no `|| echo ...`).

**Acceptance Criteria**
- `go test ./mentor/backend/src/...` passes locally with Postgres.
- CI workflow passes without suppressing failures.
- No schema changes executed during tests.
- Tests clean up via rollback; shared DB remains unchanged.

**References**
- Setup: `mentor/backend/src/database/test_db.go`
- Tests: `mentor/backend/src/database/edge_cases_test.go`, `coverage_boost_test.go`
- CI: `.github/workflows/ci.yml`

**Checklist**
- [ ] Refactor test setup for transactional isolation (no migrations in tests)
- [ ] Standardize DSN/env handling (always Postgres)
- [ ] Remove failure suppression in CI test steps
- [ ] Verify passing locally and in CI