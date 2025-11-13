# Fix Unit Tests Without Modifying Database

**Summary**
- Unit tests intermittently fail due to direct DB migrations/constraints and environment variability. The goal is to make tests reliable without modifying the live/shared database schema or data. The GORM models in `mentor/backend/src/models` are the authoritative reference for domain structure and should remain unchanged.

**Context**
- We have removed SQLite; tests should always target PostgreSQL.
- Current `SetupTestDB` attempts migrations and may drop or alter constraints, leading to failures like missing roles or constraint errors.
- CI uses a Postgres service and environment variables; locally we also use Postgres.
- Recent errors include hostname resolution issues and role-not-found when defaults are used.

**Requirements**
- Do not modify the existing database schema or constraints during tests.
- Do not change GORM model definitions (they are the source of truth).
- Make tests deterministic and independent of the current DB state.
- Tests must pass locally and in GitHub Actions using the provided Postgres service.
- Keep configuration straightforward: rely on env vars; no SQLite fallback.

**Constraints**
- No destructive operations on shared DB (e.g., dropping constraints/tables).
- Avoid global migrations in unit tests; use isolation per test.
- Do not introduce breaking changes to public APIs or model files.

**Proposed Approach**
- Replace migration-in-tests with one of the following isolation strategies:
  - Use transactions + rollback per test: begin a transaction at test setup, run operations, and `ROLLBACK` at teardown so the database state is not changed.
  - Alternatively, use a dedicated test schema (e.g., `test_schema`) created once and cleaned via transaction rollback, while leaving the main schema untouched.
  - Use GORM session with `Begin()`/`Rollback()` to keep test writes ephemeral.
- Standardize env var reading and DSN building; fail fast if required env vars are missing.
- Ensure ports and other env values are parsed and validated (e.g., `POSTGRES_PORT` via `strconv.Atoi`).
- Update GitHub Actions to ensure tests fail the job on errors (no `|| echo ...`).

**Acceptance Criteria**
- Running `go test ./mentor/backend/src/...` passes locally with a Postgres instance, without altering existing DB constraints/tables.
- Running CI workflow (`.github/workflows/ci.yml`) passes tests without ignoring failures.
- No changes to files under `mentor/backend/src/models`.
- No DDL (ALTER/DROP) statements executed during unit tests.
- Tests clean up their writes via rollback or isolated schema.

**Action Items**
1. Refactor `mentor/backend/src/database/test_db.go`:
   - Always use PostgreSQL (remove SQLite code).
   - Build DSN from env vars: `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_DB`, `SSLMODE`.
   - Parse `POSTGRES_PORT` to int; use `strconv`.
   - Remove auto-migration calls from unit test setup; do not drop/alter constraints.
   - Provide `SetupTestDBWithTx(t)` that returns a `*gorm.DB` bound to a transaction; ensure teardown rolls back.
2. Update tests under `mentor/backend/src/database/*_test.go`:
   - Use transactional setup/teardown; avoid relying on global migrations.
   - Replace any cleanup that deletes from tables with rollback.
3. CI improvements in `.github/workflows/ci.yml`:
   - Ensure test steps do not suppress failures (remove `|| echo ...`).
   - Optionally add `-count=1` to Go tests to avoid cache confusion.
4. Document required env vars in `docs/DEVELOPMENT.md` and ensure `start.sh` exports them for local runs.

**Environment Variables**
- `POSTGRES_USER`
- `POSTGRES_PASSWORD`
- `POSTGRES_HOST`
- `POSTGRES_PORT` (int)
- `POSTGRES_DB`
- `SSLMODE` (default `disable`)

**Reference Files**
- GORM models (do not change): `mentor/backend/src/models/*`
- DB setup: `mentor/backend/src/database/test_db.go`
- Tests: `mentor/backend/src/database/edge_cases_test.go`, `coverage_boost_test.go`
- CI: `.github/workflows/ci.yml`

**Notes**
- If isolation via transactions is not feasible for specific test cases (e.g., operations requiring schema existence), prefer a pre-provisioned schema created outside of unit tests (migration phase) and do not alter it during tests.
