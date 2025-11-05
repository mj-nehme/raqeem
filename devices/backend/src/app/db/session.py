from sqlalchemy.ext.asyncio import AsyncSession, create_async_engine
from sqlalchemy.orm import sessionmaker
import os
from dotenv import load_dotenv
import logging

logger = logging.getLogger(__name__)

# Load environment variables from .env file
load_dotenv()

# Get DATABASE_URL from environment with no default - will raise if not set
DATABASE_URL = os.getenv("DATABASE_URL")

# If DATABASE_URL contains shell variables like $(VAR), construct it from individual components
if DATABASE_URL and "$(POSTGRES_PASSWORD)" in DATABASE_URL:
    # Construct from individual env vars
    user = os.getenv("POSTGRES_USER", "monitor")
    password = os.getenv("POSTGRES_PASSWORD", "supersecret")
    host = os.getenv("POSTGRES_HOST", "localhost")
    port = os.getenv("POSTGRES_PORT", "5432")
    db = os.getenv("POSTGRES_DB", "monitoring_db")
    DATABASE_URL = f"postgresql+asyncpg://{user}:{password}@{host}:{port}/{db}"

if not DATABASE_URL:
    error_msg = (
        "DATABASE_URL environment variable is not set. "
        "Please set it in .env file or environment variables."
    )
    logger.error(error_msg)
    raise ValueError(error_msg)

# Don't log DATABASE_URL as it contains credentials
logger.info("Initializing database connection pool")

engine = create_async_engine(
    DATABASE_URL, 
    echo=True,
    pool_pre_ping=True,  # Enable connection health checks
    pool_size=5,
    max_overflow=10
)

async_session = sessionmaker(
    bind=engine,
    class_=AsyncSession,
    expire_on_commit=False,
)

# Dependency
async def get_db():
    """
    Database session dependency with proper error handling.
    """
    async with async_session() as session:
        try:
            yield session
        except Exception as e:
            logger.error(f"Database session error: {e}")
            await session.rollback()
            raise
        finally:
            await session.close()
