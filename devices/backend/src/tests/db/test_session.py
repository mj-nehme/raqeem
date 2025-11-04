import pytest
from app.db.session import get_db, async_session


@pytest.mark.asyncio
async def test_get_db_yields_session():
    """Test that get_db yields an async session."""
    async for session in get_db():
        assert session is not None
        # Check that it's an async session
        assert hasattr(session, 'execute')
        assert hasattr(session, 'commit')
        assert hasattr(session, 'rollback')
        break  # Only test the first yield


@pytest.mark.asyncio
async def test_get_db_context_manager():
    """Test that get_db works as a context manager."""
    generator = get_db()
    session = await generator.__anext__()
    assert session is not None
    try:
        await generator.__anext__()
    except StopAsyncIteration:
        # This is expected - generator should stop after yielding once
        pass


@pytest.mark.asyncio
async def test_database_url_configured():
    """Test that DATABASE_URL is configured."""
    import os
    assert os.getenv('DATABASE_URL') is not None


def test_async_session_configured():
    """Test that async session maker is configured."""
    assert async_session is not None
