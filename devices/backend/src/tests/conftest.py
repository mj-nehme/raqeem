# tests/conftest.py
#
# Test configuration for devices backend.
# Tests require PostgreSQL to be running. In CI, PostgreSQL is provided via
# Docker services. For local development, start PostgreSQL with:
#   docker run -d --name test-postgres -e POSTGRES_USER=monitor \
#     -e POSTGRES_PASSWORD=password -e POSTGRES_DB=monitoring_db -p 5432:5432 postgres:16

import os
import socket

import pytest
import pytest_asyncio
from sqlalchemy.ext.asyncio import async_sessionmaker, create_async_engine

# Set required environment variables for testing
TEST_ENV_VARS = {
    "DATABASE_URL": "postgresql+asyncpg://monitor:password@127.0.0.1:5432/monitoring_db",
    "MINIO_ENDPOINT": "localhost:9000",
    "MINIO_ACCESS_KEY": "minioadmin",
    "MINIO_SECRET_KEY": "minioadmin",
    "MINIO_BUCKET_NAME": "test-bucket",
    "MINIO_SECURE": "false",
    "SECRET_KEY": "test_jwt_secret_key_for_testing_purposes_only",
    "ACCESS_TOKEN_EXPIRE_MINUTES": "30",
    "MENTOR_API_URL": "http://localhost:8080",
    "REFRESH_TOKEN_EXPIRE_MINUTES": "10080",
}

# Apply test environment variables
for key, value in TEST_ENV_VARS.items():
    os.environ.setdefault(key, value)


def is_postgres_available(host: str = "127.0.0.1", port: int = 5432, timeout: float = 1.0) -> bool:
    """Check if PostgreSQL is available by attempting to connect to the port."""
    try:
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
            sock.settimeout(timeout)
            result = sock.connect_ex((host, port))
    except Exception:
        return False
    else:
        return result == 0


# Check database availability once at module load time
_db_available = is_postgres_available()


@pytest.fixture(autouse=True)
async def reset_db_engine():
    """
    Reset the database engine for each test to avoid event loop conflicts.
    This fixture ensures the global engine is disposed before tests
    and recreated in the current event loop context.

    Requires PostgreSQL to be running - tests will fail if database is unavailable.
    """
    # Import here to avoid circular imports and ensure env vars are set
    from app.db import session

    if not _db_available:
        # Database not available - yield and let individual tests handle it
        # Tests that require DB will fail, which is expected behavior
        yield
        return

    # Dispose the existing engine if it exists
    if session.engine:
        await session.engine.dispose()

    # Recreate engine in the current event loop
    session.engine = create_async_engine(TEST_ENV_VARS["DATABASE_URL"], echo=True)
    session.async_session = async_sessionmaker(
        bind=session.engine,
        expire_on_commit=False,
    )

    yield

    # Clean up after test
    await session.engine.dispose()


@pytest.fixture
def requires_db():
    """
    Fixture to ensure tests that require a database have access to PostgreSQL.
    Fails the test if PostgreSQL is not available (no skipping).
    """
    if not _db_available:
        pytest.fail(
            "PostgreSQL is not available. Tests require a running PostgreSQL instance. "
            "Start PostgreSQL with: docker run -d --name test-postgres "
            "-e POSTGRES_USER=monitor -e POSTGRES_PASSWORD=password "
            "-e POSTGRES_DB=monitoring_db -p 5432:5432 postgres:16"
        )


@pytest_asyncio.fixture(scope="function", loop_scope="function")
async def init_test_db():
    """
    Initialize database tables for tests using the test's event loop.
    This fixture creates a fresh engine in the current event loop to avoid
    'Task got Future attached to a different loop' errors.

    Requires PostgreSQL to be running.
    """
    if not _db_available:
        pytest.fail(
            "PostgreSQL is not available. Tests require a running PostgreSQL instance. "
            "Start PostgreSQL with: docker run -d --name test-postgres "
            "-e POSTGRES_USER=monitor -e POSTGRES_PASSWORD=password "
            "-e POSTGRES_DB=monitoring_db -p 5432:5432 postgres:16"
        )

    # Import here to avoid circular imports
    from app.db.base import Base

    # Create a new engine in the current event loop
    engine = create_async_engine(TEST_ENV_VARS["DATABASE_URL"], echo=False)

    try:
        async with engine.begin() as conn:
            # Create all tables
            await conn.run_sync(Base.metadata.create_all)
        yield engine
    finally:
        await engine.dispose()
