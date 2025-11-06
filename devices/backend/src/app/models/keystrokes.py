from sqlalchemy import Column, Text, TIMESTAMP
from sqlalchemy.dialects.postgresql import UUID
from app.db.base import Base
import sqlalchemy.sql as sa
import uuid

class Keystroke(Base):
    __tablename__ = "keystrokes"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    # Store user_id as plain text to avoid strict FK/UUID constraints for tests
    user_id = Column(Text, nullable=False)
    keylog = Column(Text, nullable=False)
    logged_at = Column(TIMESTAMP, server_default=sa.func.now())
