from app.db.base import Base
from app.db.session import engine
# Import models to register metadata (side-effect import to populate Base.metadata)
from app.models import users, locations, screenshots, keystrokes, devices  # noqa: F401


async def init_db():
    async with engine.begin() as conn:
        # create all tables including new device models
        await conn.run_sync(Base.metadata.create_all)
