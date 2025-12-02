---
name: Fix Unit Tests
about: Make unit tests reliable without modifying the database; keep GORM models authoritative
labels: testing, ci
---

**Summary**
- Fix all unit tests (mentor and devices Frontends and backends).
- Do NOT modify the existing database schema or constraints during tests.
- Keep GORM models in `mentor/backend/src/models` as the single source of truth.
- Check the latest Github Actions failing workflow.

**Acceptance Criteria**
- CI workflow passes without suppressing failures.
- No schema changes executed during tests.
- Tests clean up via rollback; shared DB remains unchanged.
- No tests use `@pytest.mark.skip` (policy: convert unavoidable temporary failures to `@pytest.mark.xfail` with linked issue and clear rationale; prefer fixing root cause immediately).
- New or modified tests must not introduce flakiness (prove with at least 2 consecutive green CI runs if previously flaky).
- Live dependency emulation (e.g., Postgres) uses stable session-scoped fixtures; external services (e.g., MinIO) must degrade gracefully or be explicitly mocked without skips.