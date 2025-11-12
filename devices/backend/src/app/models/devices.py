from sqlalchemy import Column, String, Float, Integer, BigInteger, TIMESTAMP, Boolean, Text
from sqlalchemy.dialects.postgresql import UUID
from app.db.base import Base
import sqlalchemy.sql as sa
import uuid


class Device(Base):
    __tablename__ = "devices"

    id = Column(String, primary_key=True)  # keep as text to allow arbitrary device ids
    name = Column(Text, nullable=True)
    type = Column(Text, nullable=True)
    os = Column(Text, nullable=True)
    last_seen = Column(TIMESTAMP, server_default=sa.func.now())
    is_online = Column(Boolean, nullable=True)
    location = Column(Text, nullable=True)
    ip_address = Column(Text, nullable=True)
    mac_address = Column(Text, nullable=True)
    current_user_text = Column(Text, nullable=True)  # Match database schema column name


class DeviceMetric(Base):
    __tablename__ = "device_metrics"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    device_id = Column(String, nullable=False)
    timestamp = Column(TIMESTAMP, server_default=sa.func.now())

    cpu_usage = Column(Float, nullable=True)
    cpu_temp = Column(Float, nullable=True)

    memory_total = Column(BigInteger, nullable=True)
    memory_used = Column(BigInteger, nullable=True)
    swap_used = Column(BigInteger, nullable=True)

    disk_total = Column(BigInteger, nullable=True)
    disk_used = Column(BigInteger, nullable=True)

    net_bytes_in = Column(BigInteger, nullable=True)
    net_bytes_out = Column(BigInteger, nullable=True)


class DeviceProcess(Base):
    __tablename__ = "device_processes"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    device_id = Column(String, nullable=False)
    timestamp = Column(TIMESTAMP, server_default=sa.func.now())
    pid = Column(Integer, nullable=False)
    name = Column(Text, nullable=False)
    cpu = Column(Float, nullable=True)
    memory = Column(BigInteger, nullable=True)
    command = Column(Text, nullable=True)


class DeviceActivity(Base):
    __tablename__ = "device_activities"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    device_id = Column(String, nullable=False)
    timestamp = Column(TIMESTAMP, server_default=sa.func.now())
    type = Column(Text, nullable=True)
    description = Column(Text, nullable=True)
    app = Column(Text, nullable=True)
    duration = Column(Integer, nullable=True)


class DeviceAlert(Base):
    __tablename__ = "device_alerts"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    device_id = Column(String, nullable=False)
    timestamp = Column(TIMESTAMP, server_default=sa.func.now())
    level = Column(Text, nullable=True)
    type = Column(Text, nullable=True)
    message = Column(Text, nullable=True)
    value = Column(Float, nullable=True)
    threshold = Column(Float, nullable=True)


class DeviceRemoteCommand(Base):
    __tablename__ = "device_remote_commands"

    id = Column(Integer, primary_key=True, autoincrement=True)
    device_id = Column(String, nullable=False)
    command = Column(Text, nullable=False)
    status = Column(Text, nullable=False)  # pending, running, completed, failed
    created_at = Column(TIMESTAMP, server_default=sa.func.now())
    completed_at = Column(TIMESTAMP, nullable=True)
    result = Column(Text, nullable=True)
    exit_code = Column(Integer, nullable=True)


class DeviceScreenshot(Base):
    __tablename__ = "device_screenshots"
    __table_args__ = {'extend_existing': True}

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    device_id = Column(String, nullable=False)
    timestamp = Column(TIMESTAMP, server_default=sa.func.now())
    path = Column(Text, nullable=False)
    resolution = Column(Text, nullable=True)
    size = Column(BigInteger, nullable=True)
