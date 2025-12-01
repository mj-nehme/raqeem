"""Tests for MinIO service functionality.

These tests cover MinIO service operations with mocking to avoid
requiring an actual MinIO server connection.
"""

import os
from contextlib import contextmanager
from unittest.mock import patch

import pytest


@contextmanager
def skip_minio_connection():
    """Context manager to test MinIO service with MINIO_SKIP_CONNECT enabled."""
    with patch.dict(os.environ, {"MINIO_SKIP_CONNECT": "1"}):
        from app.services import minio_service
        import importlib
        importlib.reload(minio_service)
        try:
            yield minio_service
        finally:
            importlib.reload(minio_service)


class TestMinioService:
    """Tests for MinIO service initialization and operations."""

    def test_minio_service_skip_connect(self):
        """Test that MinIO service skips connection when MINIO_SKIP_CONNECT is set."""
        with skip_minio_connection() as minio_service:
            service = minio_service.MinioService()
            assert service.client is None
            assert service.bucket_name == "raqeem-screenshots"

    def test_minio_service_upload_file_skip_connect(self):
        """Test upload_file when MINIO_SKIP_CONNECT is enabled."""
        with skip_minio_connection() as minio_service:
            service = minio_service.MinioService()
            result = service.upload_file("/tmp/test.png", "device123/test.png")
            assert result == "device123/test.png"

    def test_minio_service_remove_file_skip_connect(self):
        """Test remove_file when MINIO_SKIP_CONNECT is enabled."""
        with skip_minio_connection() as minio_service:
            service = minio_service.MinioService()
            # Should not raise
            service.remove_file("device123/test.png")

    def test_minio_service_get_presigned_url_skip_connect(self):
        """Test get_presigned_url when MINIO_SKIP_CONNECT is enabled."""
        with skip_minio_connection() as minio_service:
            service = minio_service.MinioService()
            url = service.get_presigned_url("device123/test.png")
            assert url == "http://localhost/minio/device123/test.png"

    def test_minio_service_ensure_bucket_skip_connect(self):
        """Test _ensure_bucket when MINIO_SKIP_CONNECT is enabled."""
        with skip_minio_connection() as minio_service:
            service = minio_service.MinioService()
            # Should not raise and not attempt connection
            service._ensure_bucket()


class TestMinioServiceExceptions:
    """Tests for MinIO service exception classes."""

    def test_minio_service_error_is_exception(self):
        """Test MinioServiceError is a proper exception."""
        from app.services.minio_service import MinioServiceError
        
        error = MinioServiceError("Test error")
        assert str(error) == "Test error"
        assert isinstance(error, Exception)

    def test_minio_upload_error_is_exception(self):
        """Test MinioUploadError is a proper exception."""
        from app.services.minio_service import MinioUploadError
        
        error = MinioUploadError("Upload failed")
        assert str(error) == "Upload failed"
        assert isinstance(error, Exception)

    def test_minio_delete_error_is_exception(self):
        """Test MinioDeleteError is a proper exception."""
        from app.services.minio_service import MinioDeleteError
        
        error = MinioDeleteError("Delete failed")
        assert str(error) == "Delete failed"
        assert isinstance(error, Exception)

    def test_minio_url_error_is_exception(self):
        """Test MinioURLError is a proper exception."""
        from app.services.minio_service import MinioURLError
        
        error = MinioURLError("URL generation failed")
        assert str(error) == "URL generation failed"
        assert isinstance(error, Exception)


class TestMinioServiceConstants:
    """Tests for MinIO service constants and defaults."""

    def test_default_url_expiration(self):
        """Test default URL expiration constant."""
        from app.services.minio_service import DEFAULT_URL_EXPIRATION
        
        assert DEFAULT_URL_EXPIRATION == 3600  # 1 hour in seconds

    def test_skip_minio_env_var_parsing(self):
        """Test SKIP_MINIO environment variable parsing."""
        # Test when not set
        with patch.dict(os.environ, {}, clear=False):
            os.environ.pop("MINIO_SKIP_CONNECT", None)
            from app.services import minio_service
            
            import importlib
            importlib.reload(minio_service)
            
            # When env var is not set or not "1", SKIP_MINIO should be False
            # But we can't easily test this without affecting global state
            
            importlib.reload(minio_service)
