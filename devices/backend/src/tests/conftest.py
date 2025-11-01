# tests/conftest.py

import os
import pytest
from dotenv import load_dotenv
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import sessionmaker
from fastapi.testclient import TestClient
from httpx import AsyncClient, ASGITransport

from app.db.base import Base  # your SQLAlchemy models metadata
from app.main import app
from app.db.session import get_db

load_dotenv()  # load .env variables

DATABASE_URL = os.getenv("DATABASE_URL")

# Create async SQLAlchemy engine and session for testing
engine = create_async_engine(DATABASE_URL, echo=False)
TestingSessionLocal = sessionmaker(
    bind=engine,
    class_=AsyncSession,
    expire_on_commit=False,
)

@pytest.fixture(scope="session")
async def prepare_database():
    # Create tables before tests run
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
    yield
    # Optional: drop tables after tests
    # async with engine.begin() as conn:
    #     await conn.run_sync(Base.metadata.drop_all)

@pytest.fixture(scope="function")
async def db_session(prepare_database):
    """Creates a new database session for a test."""
    async with TestingSessionLocal() as session:
        yield session
        # Rollback after test to keep DB clean
        await session.rollback()

# Override get_db dependency in FastAPI app
async def override_get_db():
    async with TestingSessionLocal() as session:
        yield session

app.dependency_overrides[get_db] = override_get_db

@pytest.fixture(scope="module")
def client():
    """Sync test client for FastAPI."""
    with TestClient(app) as c:
        yield c

@pytest.fixture(scope="module")
async def async_client():
    """Async test client for FastAPI."""
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        yield ac
