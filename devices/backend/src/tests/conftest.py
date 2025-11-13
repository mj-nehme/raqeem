# tests/conftest.py

import os
import pytest
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import sessionmaker
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


@pytest.fixture(scope="session")
async def engine():
    """Create a test engine for the session."""
    database_url = os.getenv('DATABASE_URL')
    test_engine = create_async_engine(database_url, echo=False)
    
    # Create all tables
    async with test_engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
    
    yield test_engine
    
    await test_engine.dispose()


@pytest.fixture(scope="function")
async def db_session(engine):
    """
    Provide a transactional database session for each test.
    All changes are rolled back after the test completes.
    """
    # Create connection
    async with engine.connect() as connection:
        # Begin transaction
        transaction = await connection.begin()
        
        # Create session bound to this connection
        async_session = sessionmaker(
            bind=connection,
            class_=AsyncSession,
            expire_on_commit=False,
        )
        
        async with async_session() as session:
            yield session
            
            # Rollback transaction - all changes are discarded
            await transaction.rollback()


@pytest.fixture(autouse=True)
def mock_dependencies():
    """Mock external dependencies for all tests."""
    # Tests use actual database with proper DATABASE_URL in CI
    # Transaction rollback ensures no data persists
    yield
