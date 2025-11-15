from app.db.base import Base
from app.db.session import engine

# Import models to register metadata (side-effect import to populate Base.metadata)
# Explicit imports ensure SQLAlchemy registers all table definitions before create_all.
from app.models import devices, users  # noqa: F401

# Simple guard to avoid repeated drop/create across tests that call init_db
_INIT_DONE = False


async def init_db():
    """Initialize database schema for tests/runtime.

    Ensures all ORM models are imported so metadata is populated, then creates
    any missing tables. This keeps test environment schema in sync with model
    definitions (resolving previous mismatches where existing legacy tables
    had differing column names).
    """
    if getattr(init_db, "_done", False):
        return
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
    init_db._done = True  # type: ignore[attr-defined]
