from sqlalchemy import Column, Text, ForeignKey, TIMESTAMP
from sqlalchemy.dialects.postgresql import UUID
from app.db.base import Base
import sqlalchemy.sql as sa
import uuid

class Keystroke(Base):
    __tablename__ = "keystrokes"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    user_id = Column(UUID(as_uuid=True), ForeignKey("users.id", ondelete="CASCADE"))
    key = Column(Text, nullable=False)
    created_at = Column(TIMESTAMP, server_default=sa.func.now())
