from sqlalchemy import Column, Text, ForeignKey, TIMESTAMP
from sqlalchemy.dialects.postgresql import UUID
from app.db.base import Base
import sqlalchemy.sql as sa
import uuid

class Screenshot(Base):
    __tablename__ = "screenshots"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    user_id = Column(UUID(as_uuid=True), ForeignKey("users.id", ondelete="CASCADE"))
    image_path = Column(Text, nullable=False)
    created_at = Column(TIMESTAMP, server_default=sa.func.now())
