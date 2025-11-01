import os
import pytest
from dotenv import load_dotenv
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy import text

load_dotenv()  # Load env vars from .env

@pytest.mark.asyncio
async def test_db_connection():
    database_url = os.getenv("DATABASE_URL")
    assert database_url is not None, "DATABASE_URL must be set in .env"

    engine = create_async_engine(database_url, echo=False)
    async with engine.connect() as conn:
        result = await conn.execute(text("SELECT 1"))
        scalar = result.scalar_one()
        assert scalar == 1
    await engine.dispose()
