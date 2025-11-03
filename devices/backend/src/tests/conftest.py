# tests/conftest.py

import os
import pytest
from unittest.mock import patch, AsyncMock

# Set required environment variables for testing
TEST_ENV_VARS = {
    'DATABASE_URL': 'postgresql://test:test@localhost:5432/test_db',
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
    with patch('app.db.database.database') as mock_db:
        mock_db.execute = AsyncMock()
        mock_db.fetch_all = AsyncMock(return_value=[])
        mock_db.fetch_one = AsyncMock(return_value=None)
        
        with patch('app.core.minio_client.minio_client') as mock_minio:
            mock_minio.put_object = AsyncMock()
            mock_minio.presigned_get_object = AsyncMock(return_value="http://test-url")
            
            yield {
                'database': mock_db,
                'minio': mock_minio
            }
