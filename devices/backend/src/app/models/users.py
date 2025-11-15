import uuid

import sqlalchemy.sql as sa
from sqlalchemy import TIMESTAMP, Column, Text
from sqlalchemy.dialects.postgresql import UUID

from app.db.base import Base


class User(Base):
    __tablename__ = "users"

    userid = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    deviceid = Column(UUID(as_uuid=True), nullable=False)
    username = Column(Text, nullable=True)
    created_at = Column(TIMESTAMP, server_default=sa.func.now())
