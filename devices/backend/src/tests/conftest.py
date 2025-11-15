# tests/conftest.py

import os

import pytest
import pytest_asyncio
from app.db.base import Base
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


@pytest.fixture(autouse=True)
async def reset_db_engine():
    """
    Reset the database engine for each test to avoid event loop conflicts.
    This fixture ensures the global engine is disposed before tests
    and recreated in the current event loop context.
    """
    # Import here to avoid circular imports and ensure env vars are set
    from app.db import session

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


@pytest_asyncio.fixture(scope="function", loop_scope="function")
async def init_test_db():
    """
    Initialize database tables for tests using the test's event loop.
    This fixture creates a fresh engine in the current event loop to avoid
    'Task got Future attached to a different loop' errors.
    """
    # Create a new engine in the current event loop
    engine = create_async_engine(TEST_ENV_VARS["DATABASE_URL"], echo=False)

    try:
        async with engine.begin() as conn:
            # Create all tables
            await conn.run_sync(Base.metadata.create_all)
        yield engine
    finally:
        await engine.dispose()
