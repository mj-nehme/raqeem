from typing import Optional
from pydantic import Field, AliasChoices
from pydantic_settings import BaseSettings, SettingsConfigDict

class Settings(BaseSettings):
    # Pydantic v2: use SettingsConfigDict instead of inner Config
    # Ignore unrelated env vars to avoid failures in diverse dev shells
    # Do NOT read .env implicitly; tests expect explicit env only
    model_config = SettingsConfigDict(
        extra="ignore",
    )

    # Database URL, e.g. postgresql+asyncpg://user:pass@localhost/dbname
    database_url: str = Field(..., validation_alias=AliasChoices("DATABASE_URL"))

    # MinIO configuration
    minio_endpoint: str = Field(..., validation_alias=AliasChoices("MINIO_ENDPOINT"))
    minio_access_key: str = Field(..., validation_alias=AliasChoices("MINIO_ACCESS_KEY"))
    minio_secret_key: str = Field(..., validation_alias=AliasChoices("MINIO_SECRET_KEY"))
    minio_secure: bool = Field(default=True, validation_alias=AliasChoices("MINIO_SECURE"))

    # Security settings
    secret_key: str = Field(..., validation_alias=AliasChoices("SECRET_KEY"))
    access_token_expire_minutes: int = Field(
        60 * 24 * 7,
        validation_alias=AliasChoices("ACCESS_TOKEN_EXPIRE_MINUTES"),
    )  # 7 days default

    # Optional: mentor backend API base URL to forward alerts/ingestion
    mentor_api_url: Optional[str] = Field(
        default=None, validation_alias=AliasChoices("MENTOR_API_URL")
    )

settings = Settings()
