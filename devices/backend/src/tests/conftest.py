# tests/conftest.py

import os
import pytest
import pytest_asyncio
from sqlalchemy.ext.asyncio import create_async_engine
from app.db.base import Base

# Set required environment variables for testing
TEST_ENV_VARS = {
    'DATABASE_URL': 'postgresql+asyncpg://monitor:password@127.0.0.1:5432/monitoring_db',
    'MINIO_ENDPOINT': 'localhost:9000',
    'MINIO_ACCESS_KEY': 'minioadmin',
    'MINIO_SECRET_KEY': 'minioadmin',
    'MINIO_BUCKET_NAME': 'test-bucket',
    'MINIO_SECURE': 'false',
    'SECRET_KEY': 'test_jwt_secret_key_for_testing_purposes_only',
    'ACCESS_TOKEN_EXPIRE_MINUTES': '30',
    'MENTOR_API_URL': 'http://localhost:8080',
    'REFRESH_TOKEN_EXPIRE_MINUTES': '10080'
}

# Apply test environment variables
for key, value in TEST_ENV_VARS.items():
    os.environ.setdefault(key, value)


@pytest.fixture(autouse=True)
def mock_dependencies():
    """
    Mock external dependencies for all tests.
    
    Note: These are integration tests that use the actual FastAPI app and PostgreSQL database.
    Tests use unique IDs to avoid conflicts. For truly non-destructive tests, consider:
    1. Using pytest-postgresql with transaction rollback
    2. Database truncation between tests
    3. Mocking the database layer entirely
    
    Current approach: Each test uses unique device IDs to prevent interference.
    """
    yield


@pytest_asyncio.fixture(scope="function")
async def init_test_db():
    """
    Initialize database tables for tests using the test's event loop.
    This fixture creates a fresh engine in the current event loop to avoid
    'Task got Future attached to a different loop' errors.
    """
    # Create a new engine in the current event loop
    engine = create_async_engine(TEST_ENV_VARS['DATABASE_URL'], echo=False)
    
    try:
        async with engine.begin() as conn:
            # Create all tables
            await conn.run_sync(Base.metadata.create_all)
        yield engine
    finally:
        await engine.dispose()
