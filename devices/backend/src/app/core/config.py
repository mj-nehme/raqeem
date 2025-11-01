from pydantic import Field
from pydantic_settings import BaseSettings
from typing import Optional

class Settings(BaseSettings):
    # Database URL, e.g. postgresql+asyncpg://user:pass@localhost/dbname
    database_url: str = Field(..., env="DATABASE_URL")

    # MinIO configuration
    minio_endpoint: str = Field(..., env="MINIO_ENDPOINT")
    minio_access_key: str = Field(..., env="MINIO_ACCESS_KEY")
    minio_secret_key: str = Field(..., env="MINIO_SECRET_KEY")
    minio_secure: bool = Field(default=True, env="MINIO_SECURE")

    # Security settings
    secret_key: str = Field(..., env="SECRET_KEY")
    access_token_expire_minutes: int = Field(60 * 24 * 7, env="ACCESS_TOKEN_EXPIRE_MINUTES")  # 7 days default

    # Optional: mentor backend API base URL to forward alerts/ingestion
    mentor_api_url: Optional[str] = Field(default=None, env="MENTOR_API_URL")

    class Config:
        env_file = ".env"
        env_file_encoding = "utf-8"

settings = Settings()
