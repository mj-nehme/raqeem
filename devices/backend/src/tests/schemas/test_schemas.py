import pytest
from uuid import uuid4
from datetime import datetime
from app.schemas.app_activity import AppActivityCreate, AppActivityOut
from app.schemas.keystrokes import KeystrokeCreate, KeystrokeOut
from app.schemas.locations import LocationCreate, LocationOut
from app.schemas.screenshots import ScreenshotCreate, ScreenshotOut
from app.schemas.users import UserCreate, UserOut


class TestAppActivitySchemas:
    """Test app activity schemas."""
    
    def test_app_activity_create(self):
        """Test creating AppActivityCreate schema."""
        user_id = uuid4()
        data = AppActivityCreate(
            user_id=user_id,
            app_name="Chrome",
            activity="opened"
        )
        assert data.user_id == user_id
        assert data.app_name == "Chrome"
        assert data.activity == "opened"
    
    def test_app_activity_out(self):
        """Test creating AppActivityOut schema."""
        user_id = uuid4()
        activity_id = uuid4()
        now = datetime.now()
        data = AppActivityOut(
            id=activity_id,
            user_id=user_id,
            app_name="Firefox",
            activity="closed",
            created_at=now
        )
        assert data.id == activity_id
        assert data.app_name == "Firefox"
        assert data.created_at == now


class TestKeystrokeSchemas:
    """Test keystroke schemas."""
    
    def test_keystroke_create(self):
        """Test creating KeystrokeCreate schema."""
        user_id = uuid4()
        data = KeystrokeCreate(
            user_id=user_id,
            key="a"
        )
        assert data.user_id == user_id
        assert data.key == "a"
    
    def test_keystroke_out(self):
        """Test creating KeystrokeOut schema."""
        user_id = uuid4()
        keystroke_id = uuid4()
        now = datetime.now()
        data = KeystrokeOut(
            id=keystroke_id,
            user_id=user_id,
            key="Enter",
            created_at=now
        )
        assert data.id == keystroke_id
        assert data.key == "Enter"
        assert data.created_at == now


class TestLocationSchemas:
    """Test location schemas."""
    
    def test_location_create(self):
        """Test creating LocationCreate schema."""
        user_id = uuid4()
        data = LocationCreate(
            user_id=user_id,
            latitude=51.5074,
            longitude=-0.1278
        )
        assert data.user_id == user_id
        assert data.latitude == 51.5074
        assert data.longitude == -0.1278
    
    def test_location_out(self):
        """Test creating LocationOut schema."""
        user_id = uuid4()
        location_id = uuid4()
        now = datetime.now()
        data = LocationOut(
            id=location_id,
            user_id=user_id,
            latitude=40.7128,
            longitude=-74.0060,
            created_at=now
        )
        assert data.id == location_id
        assert data.latitude == 40.7128
        assert data.created_at == now


class TestScreenshotSchemas:
    """Test screenshot schemas."""
    
    def test_screenshot_create(self):
        """Test creating ScreenshotCreate schema."""
        user_id = uuid4()
        data = ScreenshotCreate(
            user_id=user_id,
            image_path="path/to/screenshot.png"
        )
        assert data.user_id == user_id
        assert data.image_path == "path/to/screenshot.png"
    
    def test_screenshot_out(self):
        """Test creating ScreenshotOut schema."""
        user_id = uuid4()
        screenshot_id = uuid4()
        now = datetime.now()
        data = ScreenshotOut(
            id=screenshot_id,
            user_id=user_id,
            image_path="path/to/screenshot.png",
            created_at=now
        )
        assert data.id == screenshot_id
        assert data.image_path == "path/to/screenshot.png"
        assert data.created_at == now


class TestUserSchemas:
    """Test user schemas."""
    
    def test_user_create_with_name(self):
        """Test creating UserCreate schema with name."""
        data = UserCreate(
            device_id="device-123",
            name="John Doe"
        )
        assert data.device_id == "device-123"
        assert data.name == "John Doe"
    
    def test_user_create_without_name(self):
        """Test creating UserCreate schema without name."""
        data = UserCreate(
            device_id="device-456"
        )
        assert data.device_id == "device-456"
        assert data.name is None
    
    def test_user_out(self):
        """Test creating UserOut schema."""
        user_id = uuid4()
        now = datetime.now()
        data = UserOut(
            id=user_id,
            device_id="device-789",
            name="Jane Doe",
            created_at=now
        )
        assert data.id == user_id
        assert data.device_id == "device-789"
        assert data.name == "Jane Doe"
        assert data.created_at == now
