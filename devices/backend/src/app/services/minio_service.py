from minio import Minio
from minio.error import S3Error
from app.core.config import settings
import logging

class MinioService:
    def __init__(self):
        self.client = Minio(
            endpoint=settings.minio_endpoint,
            access_key=settings.minio_access_key,
            secret_key=settings.minio_secret_key,
            secure=settings.minio_secure,
        )

        self.bucket_name = "raqeem-screenshots"
        self._ensure_bucket()

    def _ensure_bucket(self):
        try:
            if not self.client.bucket_exists(self.bucket_name):
                self.client.make_bucket(self.bucket_name)
                logging.info(f"Created bucket: {self.bucket_name}")
        except S3Error as err:
            logging.error(f"MinIO bucket check/create failed: {err}")

    def upload_file(self, file_path: str, object_name: str) -> str:
        """
        Upload a file from local path to MinIO.
        Returns the object name (key).
        """
        try:
            self.client.fput_object(
                bucket_name=self.bucket_name,
                object_name=object_name,
                file_path=file_path,
            )
            return object_name
        except S3Error as err:
            logging.error(f"Failed to upload {object_name} to MinIO: {err}")
            raise

    def remove_file(self, object_name: str):
        try:
            self.client.remove_object(self.bucket_name, object_name)
            logging.info(f"Removed object: {object_name}")
        except S3Error as err:
            logging.error(f"Failed to remove {object_name} from MinIO: {err}")
            raise

    def get_presigned_url(self, object_name: str, expires=3600) -> str:
        """
        Generate a presigned URL for downloading the object, expires in seconds.
        """
        try:
            url = self.client.presigned_get_object(self.bucket_name, object_name, expires=expires)
            return url
        except S3Error as err:
            logging.error(f"Failed to get presigned url for {object_name}: {err}")
            raise
