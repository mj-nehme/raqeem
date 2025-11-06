from sqlalchemy import Column, TIMESTAMP, Float, Text
from sqlalchemy.dialects.postgresql import UUID
from app.db.base import Base
import sqlalchemy.sql as sa
import uuid

class Location(Base):
    __tablename__ = "locations"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    # Store user_id as plain text to allow arbitrary identifiers without FK/UUID constraints
    user_id = Column(Text, nullable=False)
    latitude = Column(Float, nullable=False)
    longitude = Column(Float, nullable=False)
    timestamp = Column(TIMESTAMP, server_default=sa.func.now())
