from sqlalchemy import Column, Text, ForeignKey, TIMESTAMP
from sqlalchemy.dialects.postgresql import UUID
from app.db.base import Base
import sqlalchemy.sql as sa
import uuid

class AppActivity(Base):
    __tablename__ = "app_activity"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    user_id = Column(UUID(as_uuid=True), ForeignKey("users.id", ondelete="CASCADE"))
    app_name = Column(Text, nullable=False)
    activity = Column(Text, nullable=False)
    created_at = Column(TIMESTAMP, server_default=sa.func.now())
