from sqlalchemy import Column, Float, Integer, BigInteger, TIMESTAMP, Boolean, Text
from sqlalchemy.dialects.postgresql import UUID
from app.db.base import Base
import sqlalchemy.sql as sa
import uuid


class Device(Base):
    __tablename__ = "devices"

    deviceid = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    device_name = Column(Text, nullable=True)
    device_type = Column(Text, nullable=True)
    os = Column(Text, nullable=True)
    last_seen = Column(TIMESTAMP, server_default=sa.func.now())
    is_online = Column(Boolean, nullable=True)
    device_location = Column(Text, nullable=True)
    ip_address = Column(Text, nullable=True)
    mac_address = Column(Text, nullable=True)
    current_user = Column(Text, nullable=True)


class DeviceMetric(Base):
    __tablename__ = "device_metrics"

    metricid = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    deviceid = Column(UUID(as_uuid=True), nullable=False)
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

    processid = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    deviceid = Column(UUID(as_uuid=True), nullable=False)
    timestamp = Column(TIMESTAMP, server_default=sa.func.now())
    pid = Column(Integer, nullable=False)
    process_name = Column(Text, nullable=False)
    cpu = Column(Float, nullable=True)
    memory = Column(BigInteger, nullable=True)
    command_text = Column(Text, nullable=True)


class DeviceActivity(Base):
    __tablename__ = "device_activities"

    activityid = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    deviceid = Column(UUID(as_uuid=True), nullable=False)
    timestamp = Column(TIMESTAMP, server_default=sa.func.now())
    activity_type = Column(Text, nullable=True)
    description = Column(Text, nullable=True)
    app = Column(Text, nullable=True)
    duration = Column(Integer, nullable=True)


class DeviceAlert(Base):
    __tablename__ = "device_alerts"

    alertid = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    deviceid = Column(UUID(as_uuid=True), nullable=False)
    timestamp = Column(TIMESTAMP, server_default=sa.func.now())
    level = Column(Text, nullable=True)
    alert_type = Column(Text, nullable=True)
    message = Column(Text, nullable=True)
    value = Column(Float, nullable=True)
    threshold = Column(Float, nullable=True)


class DeviceRemoteCommand(Base):
    __tablename__ = "device_remote_commands"

    commandid = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    deviceid = Column(UUID(as_uuid=True), nullable=False)
    command_text = Column(Text, nullable=False)
    status = Column(Text, nullable=False)
    created_at = Column(TIMESTAMP, server_default=sa.func.now())
    completed_at = Column(TIMESTAMP, nullable=True)
    result = Column(Text, nullable=True)
    exit_code = Column(Integer, nullable=True)


class DeviceScreenshot(Base):
    __tablename__ = "device_screenshots"
    __table_args__ = {'extend_existing': True}

    screenshotid = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    deviceid = Column(UUID(as_uuid=True), nullable=False)
    timestamp = Column(TIMESTAMP, server_default=sa.func.now())
    path = Column(Text, nullable=False)
    resolution = Column(Text, nullable=True)
    size = Column(BigInteger, nullable=True)
