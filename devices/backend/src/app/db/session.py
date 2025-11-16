import os

from dotenv import load_dotenv
from sqlalchemy.ext.asyncio import async_sessionmaker, create_async_engine
from sqlalchemy.pool import AsyncAdaptedQueuePool

# Load environment variables from .env file
load_dotenv()

# Get DATABASE_URL from environment with no default - will raise if not set
DATABASE_URL = os.getenv("DATABASE_URL")

# If DATABASE_URL contains shell variables like $(VAR), construct it from individual components
if DATABASE_URL and "$(POSTGRES_PASSWORD)" in DATABASE_URL:
    # Construct from individual env vars
    user = os.getenv("POSTGRES_USER", "monitor")
    password = os.getenv("POSTGRES_PASSWORD", "password")
    host = os.getenv("POSTGRES_HOST", "localhost")
    port = os.getenv("POSTGRES_PORT", "5432")
    db = os.getenv("POSTGRES_DB", "monitoring_db")
    DATABASE_URL = f"postgresql+asyncpg://{user}:{password}@{host}:{port}/{db}"

if not DATABASE_URL:
    _msg = (
        "DATABASE_URL environment variable is not set. Please set it in .env file or environment variables."
    )
    raise ValueError(_msg)

# Get connection pool settings from environment with sensible defaults
pool_size = int(os.getenv("DB_POOL_SIZE", "10"))
max_overflow = int(os.getenv("DB_MAX_OVERFLOW", "20"))
pool_timeout = int(os.getenv("DB_POOL_TIMEOUT", "30"))
pool_recycle = int(os.getenv("DB_POOL_RECYCLE", "3600"))  # 1 hour default
pool_pre_ping = os.getenv("DB_POOL_PRE_PING", "true").lower() == "true"

# Create async engine with connection pooling for production
# Use QueuePool for production, which provides better performance and reliability
engine = create_async_engine(
    DATABASE_URL,
    echo=os.getenv("DB_ECHO", "false").lower() == "true",
    poolclass=AsyncAdaptedQueuePool,
    pool_size=pool_size,
    max_overflow=max_overflow,
    pool_timeout=pool_timeout,
    pool_recycle=pool_recycle,
    pool_pre_ping=pool_pre_ping,  # Test connections before using them
)

async_session = async_sessionmaker(
    bind=engine,
    expire_on_commit=False,
)

# Backwards-compatibility for tests expecting `session.sessionmaker`
# to exist (alias to async_sessionmaker)
sessionmaker = async_sessionmaker


# Dependency
async def get_db():
    async with async_session() as session:
        yield session


# Health check function
async def health_check():
    """Check database connectivity"""
    try:
        async with engine.connect() as conn:
            await conn.execute("SELECT 1")
        return True
    except Exception:
        return False


# Shutdown function
async def shutdown():
    """Gracefully close database connections"""
    await engine.dispose()

