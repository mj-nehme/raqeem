"""API test fixtures.

Previously this module started a dedicated Postgres container per test, creating heavy
connection churn and conflicting with the session-scoped container in the root `conftest.py`.
It now intentionally does nothing so API tests reuse the shared session database.

If a future test requires explicit per-test isolation, introduce a marker and conditional
logic instead of unconditional container startup.
"""

