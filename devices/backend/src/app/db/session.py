from sqlalchemy.ext.asyncio import AsyncSession, create_async_engine
from sqlalchemy.pool import NullPool
from sqlalchemy.orm import sessionmaker
import os
from dotenv import load_dotenv

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
    raise ValueError(
        "DATABASE_URL environment variable is not set. "
        "Please set it in .env file or environment variables."
    )

# Use NullPool to avoid sharing asyncpg connections across different event loops in tests,
# which can cause "Future attached to a different loop" and concurrent operation errors.
engine = create_async_engine(DATABASE_URL, echo=True, poolclass=NullPool)

async_session = sessionmaker(
    bind=engine,
    class_=AsyncSession,
    expire_on_commit=False,
)

# Dependency
async def get_db():
    async with async_session() as session:
        yield session
