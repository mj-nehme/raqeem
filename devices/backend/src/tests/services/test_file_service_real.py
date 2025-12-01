"""Real tests for FileService - no mocking, direct function calls."""

import pytest
from app.services.file_service import FileService


class TestFileServiceReal:
    """Test FileService with real function calls."""

    @pytest.fixture
    def service(self):
        """Create a real FileService instance."""
        return FileService()

    # validate_file_type Tests
    def test_validate_file_type_png(self, service):
        """Test PNG file with correct MIME type is valid."""
        assert service.validate_file_type("image.png", "image/png") is True

    def test_validate_file_type_jpg(self, service):
        """Test JPG file with correct MIME type is valid."""
        assert service.validate_file_type("photo.jpg", "image/jpeg") is True

    def test_validate_file_type_jpeg(self, service):
        """Test JPEG file with correct MIME type is valid."""
        assert service.validate_file_type("photo.jpeg", "image/jpeg") is True

    def test_validate_file_type_gif(self, service):
        """Test GIF file with correct MIME type is valid."""
        assert service.validate_file_type("animation.gif", "image/gif") is True

    def test_validate_file_type_uppercase_extension(self, service):
        """Test uppercase file extension is valid."""
        assert service.validate_file_type("IMAGE.PNG", "image/png") is True

    def test_validate_file_type_mixed_case_extension(self, service):
        """Test mixed case file extension is valid."""
        assert service.validate_file_type("Image.JpG", "image/jpeg") is True

    def test_validate_file_type_pdf_invalid(self, service):
        """Test PDF file is invalid."""
        assert service.validate_file_type("document.pdf", "application/pdf") is False

    def test_validate_file_type_exe_invalid(self, service):
        """Test EXE file is invalid."""
        assert service.validate_file_type("script.exe", "application/exe") is False

    def test_validate_file_type_txt_invalid(self, service):
        """Test TXT file is invalid."""
        assert service.validate_file_type("data.txt", "text/plain") is False

    def test_validate_file_type_wrong_mime_for_extension(self, service):
        """Test that service allows any valid image type regardless of extension match.
        
        Note: The service only checks that BOTH are valid image types, 
        not that they match each other.
        """
        # PNG extension but JPEG MIME type - both are valid image types
        assert service.validate_file_type("image.png", "image/jpeg") is True

    def test_validate_file_type_correct_mime_wrong_extension(self, service):
        """Test correct MIME type but wrong extension is invalid."""
        # JPEG MIME type but EXE extension
        assert service.validate_file_type("image.exe", "image/jpeg") is False

    def test_validate_file_type_no_extension(self, service):
        """Test file without extension is invalid."""
        assert service.validate_file_type("image", "image/png") is False

    def test_validate_file_type_empty_filename(self, service):
        """Test empty filename is invalid."""
        assert service.validate_file_type("", "image/png") is False

    def test_validate_file_type_double_extension(self, service):
        """Test file with double extension uses the last one."""
        assert service.validate_file_type("image.txt.png", "image/png") is True

    def test_validate_file_type_webp_invalid(self, service):
        """Test WEBP file is invalid (not in supported types)."""
        assert service.validate_file_type("image.webp", "image/webp") is False

    # validate_file_size Tests
    def test_validate_file_size_small(self, service):
        """Test small file size is valid."""
        max_size = 10 * 1024 * 1024  # 10MB
        assert service.validate_file_size(1024, max_size) is True  # 1KB

    def test_validate_file_size_1mb(self, service):
        """Test 1MB file size is valid."""
        max_size = 10 * 1024 * 1024  # 10MB
        assert service.validate_file_size(1024 * 1024, max_size) is True

    def test_validate_file_size_5mb(self, service):
        """Test 5MB file size is valid."""
        max_size = 10 * 1024 * 1024  # 10MB
        assert service.validate_file_size(5 * 1024 * 1024, max_size) is True

    def test_validate_file_size_at_limit(self, service):
        """Test file at exactly the limit is valid."""
        max_size = 10 * 1024 * 1024  # 10MB
        assert service.validate_file_size(max_size, max_size) is True

    def test_validate_file_size_over_limit(self, service):
        """Test file over limit is invalid."""
        max_size = 10 * 1024 * 1024  # 10MB
        assert service.validate_file_size(11 * 1024 * 1024, max_size) is False

    def test_validate_file_size_zero(self, service):
        """Test zero file size is valid."""
        max_size = 10 * 1024 * 1024
        assert service.validate_file_size(0, max_size) is True

    def test_validate_file_size_negative(self, service):
        """Test negative file size is invalid."""
        max_size = 10 * 1024 * 1024
        assert service.validate_file_size(-1, max_size) is False

    def test_validate_file_size_large_over_limit(self, service):
        """Test very large file is invalid."""
        max_size = 10 * 1024 * 1024  # 10MB
        assert service.validate_file_size(100 * 1024 * 1024, max_size) is False

    def test_validate_file_size_very_small_limit(self, service):
        """Test with very small limit."""
        max_size = 100  # 100 bytes
        assert service.validate_file_size(50, max_size) is True
        assert service.validate_file_size(150, max_size) is False

    def test_validate_file_size_zero_limit(self, service):
        """Test with zero limit."""
        assert service.validate_file_size(0, 0) is True
        assert service.validate_file_size(1, 0) is False
