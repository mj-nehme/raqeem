"""MinIO service for S3-compatible object storage.

Provides file storage operations for screenshots and other binary data with:
- Automatic bucket creation
- Presigned URL generation for secure downloads
- Error handling with custom exceptions
- Structured logging
"""

import os
from typing import cast

from app.core.config import settings
from app.core.logging_config import get_logger
from minio import Minio
from minio.error import S3Error

logger = get_logger(__name__)

# Default expiration time for presigned URLs (1 hour)
DEFAULT_URL_EXPIRATION = 3600


class MinioServiceError(Exception):
    """Base exception for MinIO service errors."""


class MinioUploadError(MinioServiceError):
    """Exception raised when file upload fails."""


class MinioDeleteError(MinioServiceError):
    """Exception raised when file deletion fails."""


class MinioURLError(MinioServiceError):
    """Exception raised when presigned URL generation fails."""


SKIP_MINIO = os.getenv("MINIO_SKIP_CONNECT") == "1"


class MinioService:
    """Service for interacting with MinIO object storage.

    This service handles:
    - Bucket initialization and verification
    - File upload and download operations
    - Presigned URL generation for secure access
    - Error handling with custom exceptions and logging

    Example:
        >>> service = MinioService()
        >>> object_name = service.upload_file("/tmp/screenshot.png", "device123/screenshot.png")
        >>> url = service.get_presigned_url(object_name)
    """

    def __init__(self):
        """Initialize MinIO client and ensure bucket exists."""
        self.bucket_name = "raqeem-screenshots"
        if SKIP_MINIO:
            logger.info(
                "MINIO_SKIP_CONNECT=1 detected - skipping MinIO connectivity and bucket check",
                extra={"bucket": self.bucket_name},
            )
            self.client = None  # type: ignore
            return

        try:
            logger.info(
                "Initializing MinIO service",
                extra={"endpoint": settings.minio_endpoint, "secure": settings.minio_secure},
            )
            self.client = Minio(
                endpoint=settings.minio_endpoint,
                access_key=settings.minio_access_key,
                secret_key=settings.minio_secret_key,
                secure=settings.minio_secure,
            )
            logger.info("MinIO client initialized successfully")
            self._ensure_bucket()
        except Exception as e:
            logger.error(
                "Failed to initialize MinIO client",
                extra={"error": str(e), "error_type": type(e).__name__},
                exc_info=True,
            )
            msg = f"MinIO client initialization failed: {e}"
            raise MinioServiceError(msg) from e

    def _ensure_bucket(self):
        """Ensure the storage bucket exists, create if necessary.

        Raises:
            MinioServiceError: If bucket check or creation fails.
        """
        if SKIP_MINIO:
            # Short-circuit upload during tests when connectivity is skipped
            logger.debug(
                "Skipping MinIO upload (MINIO_SKIP_CONNECT=1)",
                extra={"bucket": self.bucket_name, "object_name": object_name, "file_path": file_path},
            )
            return object_name
        try:
            if not self.client.bucket_exists(self.bucket_name):
                self.client.make_bucket(self.bucket_name)
                logger.info(f"Created MinIO bucket: {self.bucket_name}")
            else:
                logger.debug(f"MinIO bucket exists: {self.bucket_name}")
        except S3Error as e:
            logger.error(
                "Failed to check/create MinIO bucket",
                extra={
                    "bucket": self.bucket_name,
                    "error": str(e),
                    "error_code": e.code if hasattr(e, "code") else "unknown",
                },
                exc_info=True,
            )
            msg = f"Failed to ensure bucket {self.bucket_name}: {e}"
            raise MinioServiceError(msg) from e

    def upload_file(self, file_path: str, object_name: str) -> str:
        """Upload a file from local path to MinIO.

        Args:
            file_path: Local filesystem path to the file to upload.
            object_name: Destination object name (key) in the bucket.

        Returns:
            The object name (key) of the uploaded file.

        Raises:
            MinioUploadError: If upload fails.

        Example:
            >>> service.upload_file("/tmp/image.png", "device123/image.png")
            "device123/image.png"
        """
        if SKIP_MINIO:
            logger.debug(
                "Skipping MinIO remove (MINIO_SKIP_CONNECT=1)",
                extra={"bucket": self.bucket_name, "object_name": object_name},
            )
            return
        try:
            logger.info(
                "Uploading file to MinIO",
                extra={
                    "bucket": self.bucket_name,
                    "object_name": object_name,
                    "file_path": file_path,
                },
            )

            self.client.fput_object(
                bucket_name=self.bucket_name,
                object_name=object_name,
                file_path=file_path,
            )

            logger.info(
                "File uploaded successfully to MinIO",
                extra={
                    "bucket": self.bucket_name,
                    "object_name": object_name,
                },
            )
            return object_name

        except S3Error as e:
            logger.error(
                "Failed to upload file to MinIO",
                extra={
                    "bucket": self.bucket_name,
                    "object_name": object_name,
                    "error": str(e),
                    "error_code": e.code if hasattr(e, "code") else "unknown",
                },
                exc_info=True,
            )
            msg = f"Failed to upload {object_name}: {e}"
            raise MinioUploadError(msg) from e

    def remove_file(self, object_name: str):
        """Remove an object from MinIO storage.

        Args:
            object_name: Object name (key) to remove.

        Raises:
            MinioDeleteError: If removal fails.

        Example:
            >>> service.remove_file("device123/image.png")
        """
        if SKIP_MINIO:
            logger.debug(
                "Skipping presigned URL generation (MINIO_SKIP_CONNECT=1)",
                extra={"bucket": self.bucket_name, "object_name": object_name, "expires": expires},
            )
            # Return deterministic placeholder URL for tests
            return f"http://localhost/minio/{object_name}"
        try:
            logger.info(
                "Removing file from MinIO",
                extra={
                    "bucket": self.bucket_name,
                    "object_name": object_name,
                },
            )

            self.client.remove_object(self.bucket_name, object_name)

            logger.info(
                "File removed successfully from MinIO",
                extra={
                    "bucket": self.bucket_name,
                    "object_name": object_name,
                },
            )

        except S3Error as e:
            logger.error(
                "Failed to remove file from MinIO",
                extra={
                    "bucket": self.bucket_name,
                    "object_name": object_name,
                    "error": str(e),
                    "error_code": e.code if hasattr(e, "code") else "unknown",
                },
                exc_info=True,
            )
            msg = f"Failed to remove {object_name}: {e}"
            raise MinioDeleteError(msg) from e

    def get_presigned_url(self, object_name: str, expires: int = DEFAULT_URL_EXPIRATION) -> str:
        """Generate a presigned URL for downloading an object.

        Args:
            object_name: Object name (key) to generate URL for.
            expires: URL expiration time in seconds (default: 1 hour).

        Returns:
            Presigned URL string for secure download.

        Raises:
            MinioURLError: If URL generation fails.

        Note:
            The URL will expire after the specified duration.

        Example:
            >>> url = service.get_presigned_url("device123/image.png", expires=7200)
            >>> # URL valid for 2 hours
        """
        try:
            logger.debug(
                "Generating presigned URL",
                extra={
                    "bucket": self.bucket_name,
                    "object_name": object_name,
                    "expires": expires,
                },
            )

            url = self.client.presigned_get_object(self.bucket_name, object_name, expires=expires)

            logger.debug(
                "Presigned URL generated successfully",
                extra={
                    "bucket": self.bucket_name,
                    "object_name": object_name,
                },
            )

            return cast("str", url)

        except S3Error as e:
            logger.error(
                "Failed to generate presigned URL",
                extra={
                    "bucket": self.bucket_name,
                    "object_name": object_name,
                    "error": str(e),
                    "error_code": e.code if hasattr(e, "code") else "unknown",
                },
                exc_info=True,
            )
            msg = f"Failed to generate presigned URL for {object_name}: {e}"
            raise MinioURLError(msg) from e
