import os

from dotenv import load_dotenv
from sqlalchemy.ext.asyncio import async_sessionmaker, create_async_engine
from sqlalchemy.pool import AsyncAdaptedQueuePool, NullPool

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
# These settings tune the connection pool for production workloads:
# - pool_size: Base number of connections maintained in the pool
# - max_overflow: Additional connections that can be created on demand
# - pool_timeout: Max seconds to wait for a connection from the pool
# - pool_recycle: Seconds after which connections are recycled (prevents stale connections)
# - pool_pre_ping: Test connections before use to detect dropped connections
pool_size = int(os.getenv("DB_POOL_SIZE", "10"))
max_overflow = int(os.getenv("DB_MAX_OVERFLOW", "20"))
pool_timeout = int(os.getenv("DB_POOL_TIMEOUT", "30"))
pool_recycle = int(os.getenv("DB_POOL_RECYCLE", "3600"))  # 1 hour default
pool_pre_ping = os.getenv("DB_POOL_PRE_PING", "true").lower() == "true"

# Use NullPool in test mode to avoid sharing connections across event loops
# NullPool creates a new connection for each request and discards it after use,
# which is safer for tests but has worse performance. In production, we use
# QueuePool which maintains a pool of persistent connections for better performance.
is_test_mode = os.getenv("PYTEST_CURRENT_TEST") is not None
poolclass = NullPool if is_test_mode else AsyncAdaptedQueuePool

# Build engine kwargs based on pool type
engine_kwargs = {
    "echo": os.getenv("DB_ECHO", "false").lower() == "true",
    "poolclass": poolclass,
}

# Only add pool parameters for non-NullPool configurations
if not is_test_mode:
    engine_kwargs.update({
        "pool_size": pool_size,
        "max_overflow": max_overflow,
        "pool_timeout": pool_timeout,
        "pool_recycle": pool_recycle,
        "pool_pre_ping": pool_pre_ping,
    })

# Create async engine with connection pooling for production
engine = create_async_engine(DATABASE_URL, **engine_kwargs)

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
    """
    Check database connectivity by executing a simple query.

    This is used by health check endpoints to verify the database is accessible.
    Returns True if the database responds successfully, False otherwise.
    Uses a simple SELECT 1 query which is fast and doesn't require any tables.
    """
    try:
        async with engine.connect() as conn:
            await conn.execute("SELECT 1")
    except Exception:
        return False
    else:
        return True


# Shutdown function
async def shutdown():
    """
    Gracefully close database connections and clean up resources.

    This should be called during application shutdown to ensure all
    database connections are properly closed and pool resources are released.
    SQLAlchemy's dispose() waits for connections to be returned to the pool
    before closing them, preventing abrupt disconnections.
    """
    await engine.dispose()

