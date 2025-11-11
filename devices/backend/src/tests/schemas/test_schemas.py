from uuid import uuid4
from datetime import datetime
from app.schemas.screenshots import ScreenshotCreate, ScreenshotOut
from app.schemas.users import UserCreate, UserOut


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
