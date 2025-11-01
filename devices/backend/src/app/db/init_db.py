from app.db.base import Base
from app.db.session import engine
from app.models import user, location, screenshot, keystroke, app_activity, devices


async def init_db():
    async with engine.begin() as conn:
        # create all tables including new device models
        await conn.run_sync(Base.metadata.create_all)
