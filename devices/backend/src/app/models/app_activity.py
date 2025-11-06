from sqlalchemy import Column, Text, TIMESTAMP, CheckConstraint
from sqlalchemy.dialects.postgresql import UUID
from app.db.base import Base
import sqlalchemy.sql as sa
import uuid

class AppActivity(Base):
    __tablename__ = "app_activity"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    # Store user_id as plain text to allow external identifiers in tests without strict FK/UUID constraints
    user_id = Column(Text, nullable=False)
    app_name = Column(Text, nullable=False)
    action = Column(Text, nullable=False, server_default='open')
    activity_time = Column(TIMESTAMP, server_default=sa.func.now())
    
    __table_args__ = (
        CheckConstraint("action IN ('open', 'close', 'background')", name='check_action_values'),
    )
