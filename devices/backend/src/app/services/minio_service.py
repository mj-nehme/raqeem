import logging
from typing import cast

from app.core.config import settings
from minio import Minio
from minio.error import S3Error

# Configure logger for this module
logger = logging.getLogger(__name__)

# Default expiration time for presigned URLs (1 hour)
DEFAULT_URL_EXPIRATION = 3600


class MinioServiceError(Exception):
    """Base exception for MinIO service errors."""

    pass


class MinioUploadError(MinioServiceError):
    """Exception raised when file upload fails."""

    pass


class MinioDeleteError(MinioServiceError):
    """Exception raised when file deletion fails."""

    pass


class MinioURLError(MinioServiceError):
    """Exception raised when presigned URL generation fails."""

    pass


class MinioService:
    """Service for managing file storage operations with MinIO/S3.

    Handles bucket creation, file uploads, deletions, and presigned URL generation.
    """

    def __init__(self):
        """Initialize MinIO client and ensure bucket exists."""
        try:
            self.client = Minio(
                endpoint=settings.minio_endpoint,
                access_key=settings.minio_access_key,
                secret_key=settings.minio_secret_key,
                secure=settings.minio_secure,
            )
            logger.info("MinIO client initialized successfully")
        except Exception as e:
            logger.exception("Failed to initialize MinIO client")
            raise MinioServiceError(f"MinIO client initialization failed: {e}") from e

        self.bucket_name = "raqeem-screenshots"
        self._ensure_bucket()

    def _ensure_bucket(self):
        """Ensure the required bucket exists, create if it doesn't.

        Raises:
            MinioServiceError: If bucket check or creation fails
        """
        try:
            if not self.client.bucket_exists(self.bucket_name):
                self.client.make_bucket(self.bucket_name)
                logger.info("Created bucket: %s", self.bucket_name)
            else:
                logger.debug("Bucket %s already exists", self.bucket_name)
        except S3Error as e:
            logger.exception("MinIO bucket check/create failed for bucket: %s", self.bucket_name)
            raise MinioServiceError(f"Failed to ensure bucket {self.bucket_name}: {e}") from e

    def upload_file(self, file_path: str, object_name: str) -> str:
        """Upload a file from local path to MinIO.

        Args:
            file_path: Local filesystem path to the file to upload
            object_name: Object key/name in the bucket

        Returns:
            The object name (key) of the uploaded file

        Raises:
            MinioUploadError: If upload operation fails
        """
        try:
            self.client.fput_object(
                bucket_name=self.bucket_name,
                object_name=object_name,
                file_path=file_path,
            )
            logger.info("Successfully uploaded file to MinIO: %s", object_name)
            return object_name
        except S3Error as e:
            logger.exception("Failed to upload %s to MinIO", object_name)
            raise MinioUploadError(f"Failed to upload {object_name}: {e}") from e

    def remove_file(self, object_name: str):
        """Remove a file from MinIO storage.

        Args:
            object_name: Object key/name in the bucket to remove

        Raises:
            MinioDeleteError: If deletion operation fails
        """
        try:
            self.client.remove_object(self.bucket_name, object_name)
            logger.info("Removed object from MinIO: %s", object_name)
        except S3Error as e:
            logger.exception("Failed to remove %s from MinIO", object_name)
            raise MinioDeleteError(f"Failed to remove {object_name}: {e}") from e

    def get_presigned_url(self, object_name: str, expires: int = DEFAULT_URL_EXPIRATION) -> str:
        """Generate a presigned URL for downloading an object.

        Args:
            object_name: Object key/name in the bucket
            expires: URL expiration time in seconds (default: 3600 / 1 hour)

        Returns:
            Presigned URL string for accessing the object

        Raises:
            MinioURLError: If presigned URL generation fails

        Note:
            The URL will expire after the specified duration
        """
        try:
            url = self.client.presigned_get_object(self.bucket_name, object_name, expires=expires)
            logger.debug("Generated presigned URL for %s (expires in %ds)", object_name, expires)
            return cast("str", url)
        except S3Error as e:
            logger.exception("Failed to get presigned URL for %s", object_name)
            raise MinioURLError(f"Failed to generate presigned URL for {object_name}: {e}") from e
