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