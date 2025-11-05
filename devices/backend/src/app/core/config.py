from typing import Optional
from pydantic import Field, AliasChoices, field_validator
from pydantic_settings import BaseSettings, SettingsConfigDict
import os


class Settings(BaseSettings):
    # Pydantic v2: use SettingsConfigDict instead of inner Config
    model_config = SettingsConfigDict(
        env_file=".env", 
        env_file_encoding="utf-8",
        extra="ignore"  # Ignore extra environment variables
    )

    # Database URL, e.g. postgresql+asyncpg://user:pass@localhost/dbname
    database_url: str = Field(..., validation_alias=AliasChoices("DATABASE_URL"))

    # MinIO configuration - can be set via MINIO_ENDPOINT or constructed from MINIO_HOST + MINIO_PORT
    minio_endpoint: Optional[str] = Field(
        default=None, 
        validation_alias=AliasChoices("MINIO_ENDPOINT")
    )
    minio_access_key: str = Field(
        default="minioadmin", 
        validation_alias=AliasChoices("MINIO_ACCESS_KEY")
    )
    minio_secret_key: str = Field(
        default="minioadmin", 
        validation_alias=AliasChoices("MINIO_SECRET_KEY")
    )
    minio_secure: bool = Field(
        default=False, 
        validation_alias=AliasChoices("MINIO_SECURE")
    )

    # Security settings
    secret_key: str = Field(
        default="dev-secret-key-change-in-production",
        validation_alias=AliasChoices("SECRET_KEY")
    )
    access_token_expire_minutes: int = Field(
        60 * 24 * 7,
        validation_alias=AliasChoices("ACCESS_TOKEN_EXPIRE_MINUTES"),
    )  # 7 days default

    # Optional: mentor backend API base URL to forward alerts/ingestion
    mentor_api_url: Optional[str] = Field(
        default=None, validation_alias=AliasChoices("MENTOR_API_URL")
    )
    
    def __init__(self, **kwargs):
        super().__init__(**kwargs)
        # Build MINIO_ENDPOINT from MINIO_HOST and MINIO_PORT if not provided
        if not self.minio_endpoint:
            minio_host = os.getenv("MINIO_HOST", "localhost")
            minio_port = os.getenv("MINIO_PORT", "9000")
            self.minio_endpoint = f"{minio_host}:{minio_port}"


settings = Settings()
