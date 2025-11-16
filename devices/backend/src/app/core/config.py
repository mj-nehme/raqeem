"""Application configuration management using Pydantic Settings."""

from pydantic import AliasChoices, Field, field_validator
from pydantic_settings import BaseSettings, SettingsConfigDict


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

    # MinIO configuration
    minio_endpoint: str = Field(
        ...,
        validation_alias=AliasChoices("MINIO_ENDPOINT"),
        description="MinIO/S3 server endpoint (host:port or with protocol for compatibility)",
    )
    minio_access_key: str = Field(
        ...,
        validation_alias=AliasChoices("MINIO_ACCESS_KEY"),
        description="MinIO/S3 access key for authentication",
    )
    minio_secret_key: str = Field(
        ...,
        validation_alias=AliasChoices("MINIO_SECRET_KEY"),
        description="MinIO/S3 secret key for authentication",
    )
    minio_secure: bool = Field(
        default=True,
        validation_alias=AliasChoices("MINIO_SECURE"),
        description="Use HTTPS for MinIO connections (default: True)",
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
            raise ValueError(
                "DATABASE_URL must use asyncpg driver. "
                "Expected format: postgresql+asyncpg://user:pass@host:port/dbname"
            )
        return v

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
        if len(v) < 32:
            # Allow short keys for testing but log warning
            import logging

            logging.warning(
                "SECRET_KEY is shorter than recommended 32 characters. "
                "This is acceptable for testing but should be at least 32 characters in production."
            )
        return v


# Global settings instance
# Type ignore is needed because Pydantic handles field validation at runtime
settings = Settings()  # type: ignore[call-arg]
