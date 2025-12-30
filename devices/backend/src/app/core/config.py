"""Application configuration management using Pydantic Settings."""

import logging
from urllib.parse import urlparse

from pydantic import AliasChoices, Field, field_validator
from pydantic_settings import BaseSettings, SettingsConfigDict

# Configuration constants
MIN_SECRET_KEY_LENGTH = 32

# Configure logger
logger = logging.getLogger(__name__)


class Settings(BaseSettings):
    """Application settings with environment variable validation.

    All settings are loaded from environment variables with validation.
    Required fields will raise ValidationError if not provided.
    """

    # Pydantic v2: use SettingsConfigDict instead of inner Config
    # Ignore unrelated env vars to avoid failures in diverse dev shells
    # Do NOT read .env implicitly; tests expect explicit env only
    model_config = SettingsConfigDict(
        extra="ignore",
    )

    # Database URL, e.g. postgresql+asyncpg://user:pass@localhost/dbname
    database_url: str = Field(
        ...,
        validation_alias=AliasChoices("DATABASE_URL"),
        description="PostgreSQL database connection URL with asyncpg driver",
    )

    # MinIO/S3-compatible object storage configuration
    minio_endpoint: str = Field(
        ...,
        validation_alias=AliasChoices("BUCKET_ENDPOINT", "BUCKET_ENDPOINT"),
        description="S3/MinIO endpoint (host:port or with protocol for compatibility)",
    )
    minio_access_key: str | None = Field(
        default=None,
        validation_alias=AliasChoices("BUCKET_ACCESS_KEY", "BUCKET_ACCESS_KEY"),
        description="Access key for S3/MinIO (omit when using IAM/IRSA)",
    )
    minio_secret_key: str | None = Field(
        default=None,
        validation_alias=AliasChoices("BUCKET_SECRET_KEY", "BUCKET_SECRET_KEY"),
        description="Secret key for S3/MinIO (omit when using IAM/IRSA)",
    )
    minio_secure: bool = Field(
        default=True,
        validation_alias=AliasChoices("BUCKET_SECURE", "BUCKET_SECURE"),
        description="Use HTTPS for S3/MinIO connections (default: True)",
    )
    minio_bucket_name: str = Field(
        default="raqeem-screenshots",
        validation_alias=AliasChoices("BUCKET_NAME", "BUCKET_NAME"),
        description="Bucket name for storing screenshots and artifacts",
    )

    # Security settings
    secret_key: str = Field(
        ...,
        validation_alias=AliasChoices("SECRET_KEY"),
        description="Secret key for JWT token signing and cryptographic operations",
    )
    access_token_expire_minutes: int = Field(
        60 * 24 * 7,  # 7 days default
        validation_alias=AliasChoices("ACCESS_TOKEN_EXPIRE_MINUTES"),
        description="JWT access token expiration time in minutes (default: 7 days)",
        ge=1,
    )

    # Optional: mentor backend API base URL to forward alerts/ingestion
    mentor_api_url: str | None = Field(
        default=None,
        validation_alias=AliasChoices("MENTOR_API_URL"),
        description="Base URL of mentor backend API for forwarding data (optional)",
    )

    @field_validator("database_url")
    @classmethod
    def validate_database_url(cls, v: str) -> str:
        """Validate database URL format and driver.

        Args:
            v: Database URL string

        Returns:
            Validated database URL

        Raises:
            ValueError: If URL doesn't use asyncpg driver
        """
        # Strip whitespace for better error messages
        v = v.strip()
        if not v.startswith("postgresql+asyncpg://"):
            msg = (
                "DATABASE_URL must use asyncpg driver. "
                "Expected format: postgresql+asyncpg://user:pass@host:port/dbname"
            )
            raise ValueError(msg)
        return v

    @field_validator("minio_endpoint")
    @classmethod
    def validate_minio_endpoint(cls, v: str) -> str:
        """Validate and sanitize MinIO endpoint format.

        Args:
            v: MinIO endpoint string (may include protocol and/or path)

        Returns:
            Sanitized endpoint in host:port format

        Raises:
            ValueError: If endpoint contains a path component (not supported by MinIO client)

        Note:
            The MinIO Python client expects endpoint in 'host:port' format without protocol.
            The protocol (HTTP/HTTPS) is determined by the 'secure' parameter.
            This validator strips any protocol prefix but rejects paths as they are not supported.
        """
        # Strip whitespace
        v = v.strip()

        # Parse the endpoint to extract components
        # If no scheme is present, add one temporarily for parsing
        parsed = urlparse(v if "://" in v else f"http://{v}")

        # Check for path component (not supported by MinIO client)
        if parsed.path and parsed.path != "/":
            msg = (
                f"BUCKET_ENDPOINT/BUCKET_ENDPOINT cannot contain a path component ('{parsed.path}'). "
                "MinIO client only supports 'host:port' format. "
                "Please remove the path from the endpoint."
            )
            raise ValueError(msg)

        # Extract the netloc (host:port) without protocol
        # If netloc is present (valid URL), use it; otherwise use original value
        endpoint = parsed.netloc if parsed.netloc else v

        # Log if we stripped the protocol for transparency
        if v != endpoint:
            logger.info(
                "Sanitized BUCKET_ENDPOINT by removing protocol: '%s' -> '%s'",
                v,
                endpoint,
            )

        return endpoint

    @field_validator("secret_key")
    @classmethod
    def validate_secret_key(cls, v: str) -> str:
        """Validate secret key length.

        Args:
            v: Secret key string

        Returns:
            Validated secret key

        Note:
            Warns in production if key is too short but allows for testing
        """
        # Strip whitespace
        v = v.strip()
        if len(v) < MIN_SECRET_KEY_LENGTH:
            # Allow short keys for testing but log warning
            logger.warning(
                "SECRET_KEY is shorter than recommended %d characters. "
                "This is acceptable for testing but should be at least %d characters in production.",
                MIN_SECRET_KEY_LENGTH,
                MIN_SECRET_KEY_LENGTH,
            )
        return v


# Global settings instance
# Type ignore is needed because Pydantic handles field validation at runtime
settings = Settings()  # type: ignore[call-arg]
