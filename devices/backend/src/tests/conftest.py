# tests/conftest.py

import os
import socket
from pathlib import Path

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

# List of test files that are known to NOT require database
# (tests that mock dependencies or don't use database at all)
_NO_DB_REQUIRED_FILES = [
    "test_main.py",
    "test_devices_simple.py",
    "test_coverage_boost.py",
    "test_endpoints_mvp_coverage.py",
    "test_error_handling.py",
    "test_legacy_field_rejection.py",
    "test_service_coverage_mvp.py",
    "test_screenshots.py",  # Has its own mocks
    "test_health.py",
    "test_minio_service.py",
    "test_http_retry.py",
    "test_logging_config.py",
]

# List of test files/classes that definitely need database
_DB_REQUIRED_PATTERNS = [
    "test_devices.py",
    "test_device_forwarding.py",
    "test_comprehensive_endpoints.py",
    "test_alerts_forwarding.py",
    "test_screenshot_forwarding.py",
    "test_init_db.py",
    "test_config.py::test_db_connection",
]


def pytest_collection_modifyitems(config, items):
    """Skip database-dependent tests when PostgreSQL is not available."""
    if _db_available:
        # Database is available, run all tests normally
        return

    skip_db = pytest.mark.skip(reason="PostgreSQL is not available - skipping integration test")

    for item in items:
        # Get the test file name using pathlib for consistent path handling
        test_file = Path(item.fspath).name
        test_nodeid = item.nodeid

        # Check if this test file is known to NOT require database
        if test_file in _NO_DB_REQUIRED_FILES:
            continue

        # Check if this test matches a database-required pattern
        for pattern in _DB_REQUIRED_PATTERNS:
            if pattern in test_nodeid or test_file == pattern:
                item.add_marker(skip_db)
                break
        else:
            # For unknown tests, check if they're in db/ folder or certain api folders
            if '/db/' in test_nodeid or 'test_db_' in test_nodeid:
                item.add_marker(skip_db)
            # Also skip comprehensive API tests that do database operations
            elif '/api/' in test_nodeid and test_file not in _NO_DB_REQUIRED_FILES:
                # Check if it's a service test or other unit test
                if '/services/' in test_nodeid:
                    # Service unit tests don't need DB
                    continue
                # Skip other API tests that are integration tests
                item.add_marker(skip_db)


@pytest.fixture(autouse=True)
async def reset_db_engine():
    """
    Reset the database engine for each test to avoid event loop conflicts.
    This fixture ensures the global engine is disposed before tests
    and recreated in the current event loop context.

    NOTE: Skips database operations if PostgreSQL is not available.
    """
    # Import here to avoid circular imports and ensure env vars are set
    from app.db import session

    if not _db_available:
        # Skip database operations if PostgreSQL is not running
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
    """Fixture to skip tests that require a database when PostgreSQL is not available."""
    if not _db_available:
        pytest.skip("PostgreSQL is not available - skipping integration test")


@pytest_asyncio.fixture(scope="function", loop_scope="function")
async def init_test_db():
    """
    Initialize database tables for tests using the test's event loop.
    This fixture creates a fresh engine in the current event loop to avoid
    'Task got Future attached to a different loop' errors.

    NOTE: Skips if PostgreSQL is not available.
    """
    if not _db_available:
        pytest.skip("PostgreSQL is not available - skipping database initialization")

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
