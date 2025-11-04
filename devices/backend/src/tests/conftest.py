# tests/conftest.py

import os
import pytest
from unittest.mock import patch, AsyncMock

# Set required environment variables for testing
TEST_ENV_VARS = {
    'DATABASE_URL': 'postgresql+asyncpg://test:test@localhost:5432/test_db',
    'MINIO_ENDPOINT': 'localhost:9000',
    'MINIO_ACCESS_KEY': 'test_access_key',
    'MINIO_SECRET_KEY': 'test_secret_key',
    'MINIO_BUCKET_NAME': 'test-bucket',
    'JWT_SECRET_KEY': 'test_jwt_secret_key_for_testing_purposes_only',
    'ACCESS_TOKEN_EXPIRE_MINUTES': '30',
    'MENTOR_BACKEND_URL': 'http://localhost:8080',
    'REFRESH_TOKEN_EXPIRE_MINUTES': '10080'
}

# Apply test environment variables
for key, value in TEST_ENV_VARS.items():
    os.environ.setdefault(key, value)

@pytest.fixture(autouse=True)
def mock_dependencies():
    """Mock external dependencies for all tests."""
    # Tests use actual database with proper DATABASE_URL in CI
    # No need to mock database or minio since tests run with proper environment
    yield
