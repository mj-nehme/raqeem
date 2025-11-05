import pytest
from app.db.base import Base


def test_base_exists():
    """Test that declarative base exists."""
    assert Base is not None


def test_base_has_metadata():
    """Test that base has metadata."""
    assert hasattr(Base, 'metadata')
    assert Base.metadata is not None


def test_base_metadata_tables():
    """Test that base metadata has tables."""
    # After models are imported, metadata should have tables
    assert len(Base.metadata.tables) > 0
