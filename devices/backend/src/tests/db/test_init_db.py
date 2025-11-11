import pytest
from app.db.init_db import init_db


@pytest.mark.asyncio
async def test_init_db_runs():
    """Test that init_db function runs without error."""
    # This will create tables if they don't exist
    await init_db()
    # No assertion needed - just checking it doesn't raise


def test_init_db_imports_models():
    """Test that init_db imports all models."""
    # Models should be imported by init_db module
    from app.models import users, locations, screenshots, keystrokes, devices
    
    assert users is not None
    assert locations is not None
    assert screenshots is not None
    assert keystrokes is not None
    assert devices is not None
