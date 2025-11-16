import logging
from typing import cast

from app.core.config import settings
from app.core.reliability import RetryConfig, retry_with_backoff, external_service_retry_config
from minio import Minio
from minio.error import S3Error


class MinioService:
    def __init__(self):
        # Use retry logic for MinIO initialization
        def init_client():
            self.client = Minio(
                endpoint=settings.minio_endpoint,
                access_key=settings.minio_access_key,
                secret_key=settings.minio_secret_key,
                secure=settings.minio_secure,
            )
            # Test connection by listing buckets
            list(self.client.list_buckets())
            return self.client

        try:
            retry_with_backoff(
                external_service_retry_config(),
                init_client,
                "MinIO initialization"
            )
        except Exception as e:
            logging.warning(f"MinIO initialization failed after retries: {e}")
            # Don't fail completely, allow service to start
            self.client = None

        self.bucket_name = "raqeem-screenshots"
        if self.client:
            self._ensure_bucket()

    def _ensure_bucket(self):
        if not self.client:
            logging.warning("MinIO client not initialized, skipping bucket check")
            return

        try:
            if not self.client.bucket_exists(self.bucket_name):
                self.client.make_bucket(self.bucket_name)
                logging.info(f"Created bucket: {self.bucket_name}")
        except S3Error:
            logging.exception("MinIO bucket check/create failed")

    def upload_file(self, file_path: str, object_name: str) -> str:
        """
        Upload a file from local path to MinIO.
        Returns the object name (key).
        """
        if not self.client:
            raise RuntimeError("MinIO client not initialized")

        try:
            self.client.fput_object(
                bucket_name=self.bucket_name,
                object_name=object_name,
                file_path=file_path,
            )
        except S3Error:
            logging.exception("Failed to upload %s to MinIO", object_name)
            raise
        else:
            return object_name

    def remove_file(self, object_name: str):
        if not self.client:
            raise RuntimeError("MinIO client not initialized")

        try:
            self.client.remove_object(self.bucket_name, object_name)
            logging.info(f"Removed object: {object_name}")
        except S3Error:
            logging.exception("Failed to remove %s from MinIO", object_name)
            raise

    def get_presigned_url(self, object_name: str, expires: int = 3600) -> str:
        """
        Generate a presigned URL for downloading the object, expires in seconds.
        """
        if not self.client:
            raise RuntimeError("MinIO client not initialized")

        try:
            url = self.client.presigned_get_object(self.bucket_name, object_name, expires=expires)
            return cast("str", url)
        except S3Error:
            logging.exception("Failed to get presigned url for %s", object_name)
            raise


async def check_minio_health() -> dict:
    """Check MinIO connection health."""
    try:
        from app.core.config import settings
        client = Minio(
            endpoint=settings.minio_endpoint,
            access_key=settings.minio_access_key,
            secret_key=settings.minio_secret_key,
            secure=settings.minio_secure,
        )
        # List buckets as a health check
        list(client.list_buckets())
        return {"status": "healthy"}
    except Exception as e:
        logging.error(f"MinIO health check failed: {e}")
        return {"status": "unhealthy", "error": str(e)}
