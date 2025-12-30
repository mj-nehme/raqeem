"""Test configuration for devices backend.

Provides a single session-scoped Postgres testcontainer and async SQLAlchemy engine.
Avoids per-test engine churn and event loop cross-talk. No test skips: failures indicate
environment or logic issues that must be addressed.
"""

import asyncio
import logging
import os
import socket
import time

import pytest
import pytest_asyncio
from sqlalchemy.ext.asyncio import async_sessionmaker, create_async_engine

# Static environment variables (non-DB) for deterministic test behavior
STATIC_TEST_ENV_VARS = {
    "BUCKET_ENDPOINT": "localhost:9000",
    "BUCKET_ACCESS_KEY": "minioadmin",
    "BUCKET_SECRET_KEY": "minioadmin",
    "BUCKET_NAME": "test-bucket",
    "BUCKET_SECURE": "false",
    "SECRET_KEY": "test_jwt_secret_key_for_testing_purposes_only",
    "ACCESS_TOKEN_EXPIRE_MINUTES": "30",
    "MENTOR_API_URL": "http://localhost:8080",
    "REFRESH_TOKEN_EXPIRE_MINUTES": "10080",
}
for key, value in STATIC_TEST_ENV_VARS.items():
    os.environ.setdefault(key, value)

# Disable Ryuk reaper to reduce port churn in constrained environments
os.environ.setdefault("TESTCONTAINERS_RYUK_DISABLED", "true")


@pytest.fixture(scope="session", autouse=True)
def _start_postgres_container():
    """Start a single Postgres testcontainer and initialize schema once for the session."""
    from app.db import session as db_session
    from app.db.base import Base
    from testcontainers.postgres import PostgresContainer

    logging.getLogger(__name__).info("Starting Postgres testcontainer (session-scoped)...")
    container = PostgresContainer(
        "postgres:16",
        username="monitor",
        password="password",
        dbname="monitoring_db",
    )
    container.start()

    try:
        host = getattr(container, "get_container_host_ip", lambda: "127.0.0.1")()
        try:
            port = int(container.get_exposed_port(5432))
        except Exception:
            port = 5432

        # Wait for port readiness (simple TCP poll)
        deadline = time.time() + 30
        last_err = None
        while time.time() < deadline:
            try:
                with socket.create_connection((host, port), timeout=1):
                    break
            except OSError as e:
                last_err = e
                time.sleep(0.2)
        else:
            msg = f"Postgres did not open port {host}:{port}: {last_err}"
            raise RuntimeError(msg)

        async_url = f"postgresql+asyncpg://monitor:password@{host}:{port}/monitoring_db"
        os.environ["DATABASE_URL"] = async_url

        from sqlalchemy.pool import NullPool
        db_session.engine = create_async_engine(async_url, echo=False, poolclass=NullPool)
        db_session.async_session = async_sessionmaker(bind=db_session.engine, expire_on_commit=False)

        async def _init_schema():
            async with db_session.engine.begin() as conn:
                await conn.run_sync(Base.metadata.create_all)

        asyncio.run(_init_schema())
        yield
    finally:
        # Cleanup: dispose engine then stop container (ignore errors)
        try:
            asyncio.run(db_session.engine.dispose())
        except Exception:
            pass
        try:
            container.stop()
        except Exception:
            pass


@pytest_asyncio.fixture()
async def db_session_scope():
    """Yield the async session factory for tests needing DB access."""
    from app.db import session as db_session
    yield db_session.async_session


@pytest_asyncio.fixture()
async def init_test_db():
    """Backward-compatible no-op fixture (schema already initialized)."""
    yield
