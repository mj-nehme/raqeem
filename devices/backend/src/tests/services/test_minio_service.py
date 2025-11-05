"""Tests for MinIO service."""
import pytest
from unittest.mock import Mock, patch
from app.services.minio_service import MinioService
from minio.error import S3Error


class TestMinioService:
    """Test MinIO service."""
    
    @patch('app.services.minio_service.Minio')
    def test_init_creates_client(self, mock_minio):
        """Test that __init__ creates a Minio client."""
        mock_client = Mock()
        mock_minio.return_value = mock_client
        mock_client.bucket_exists.return_value = True
        
        service = MinioService()
        
        assert service.client == mock_client
        assert service.bucket_name == "raqeem-screenshots"
        mock_minio.assert_called_once()
    
    @patch('app.services.minio_service.Minio')
    def test_ensure_bucket_creates_if_not_exists(self, mock_minio):
        """Test that bucket is created if it doesn't exist."""
        mock_client = Mock()
        mock_minio.return_value = mock_client
        mock_client.bucket_exists.return_value = False
        
        service = MinioService()
        
        mock_client.make_bucket.assert_called_once_with("raqeem-screenshots")
    
    @patch('app.services.minio_service.Minio')
    def test_ensure_bucket_does_not_create_if_exists(self, mock_minio):
        """Test that bucket is not created if it exists."""
        mock_client = Mock()
        mock_minio.return_value = mock_client
        mock_client.bucket_exists.return_value = True
        
        service = MinioService()
        
        mock_client.make_bucket.assert_not_called()
    
    @patch('app.services.minio_service.Minio')
    def test_upload_file_success(self, mock_minio):
        """Test successful file upload."""
        mock_client = Mock()
        mock_minio.return_value = mock_client
        mock_client.bucket_exists.return_value = True
        
        service = MinioService()
        result = service.upload_file("/path/to/file.png", "object123.png")
        
        assert result == "object123.png"
        mock_client.fput_object.assert_called_once_with(
            bucket_name="raqeem-screenshots",
            object_name="object123.png",
            file_path="/path/to/file.png"
        )
    
    @patch('app.services.minio_service.Minio')
    def test_remove_file_success(self, mock_minio):
        """Test successful file removal."""
        mock_client = Mock()
        mock_minio.return_value = mock_client
        mock_client.bucket_exists.return_value = True
        
        service = MinioService()
        service.remove_file("object123.png")
        
        mock_client.remove_object.assert_called_once_with("raqeem-screenshots", "object123.png")
    
    @patch('app.services.minio_service.Minio')
    def test_get_presigned_url_success(self, mock_minio):
        """Test getting presigned URL."""
        mock_client = Mock()
        mock_minio.return_value = mock_client
        mock_client.bucket_exists.return_value = True
        mock_client.presigned_get_object.return_value = "https://minio.example.com/presigned-url"
        
        service = MinioService()
        url = service.get_presigned_url("object123.png", expires=7200)
        
        assert url == "https://minio.example.com/presigned-url"
        mock_client.presigned_get_object.assert_called_once_with(
            "raqeem-screenshots", "object123.png", expires=7200
        )
