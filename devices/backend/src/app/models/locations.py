from sqlalchemy import Column, ForeignKey, TIMESTAMP, Double
from sqlalchemy.dialects.postgresql import UUID
from app.db.base import Base
import sqlalchemy.sql as sa
import uuid

class Location(Base):
    __tablename__ = "locations"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    user_id = Column(UUID(as_uuid=True), ForeignKey("users.id", ondelete="CASCADE"))
    latitude = Column(Double, nullable=False)
    longitude = Column(Double, nullable=False)
    created_at = Column(TIMESTAMP, server_default=sa.func.now())
