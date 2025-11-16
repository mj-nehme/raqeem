"""File service for validating file types and sizes."""

from typing import ClassVar


class FileService:
    """Service for file validation operations.

    This is a stub service primarily used for testing.
    Methods can be patched in tests to customize behavior.
    """

    IMAGE_TYPES: ClassVar[set[str]] = {"image/png", "image/jpeg", "image/gif"}

    def validate_file_type(self, filename: str, content_type: str) -> bool:
        """Validate that a file has an allowed image type.

        Checks both the MIME type and file extension to ensure consistency.

        Args:
            filename: The name of the file including extension
            content_type: The MIME type of the file (e.g., 'image/png')

        Returns:
            True if both content type and extension are valid image types, False otherwise

        Note:
            Supported image types: PNG, JPEG, and GIF
        """
        return content_type in self.IMAGE_TYPES and any(
            filename.lower().endswith(ext) for ext in [".png", ".jpg", ".jpeg", ".gif"]
        )

    def validate_file_size(self, size: int, max_size: int) -> bool:
        """Validate that a file size is within allowed limits.

        Args:
            size: The size of the file in bytes
            max_size: The maximum allowed file size in bytes

        Returns:
            True if file size is between 0 and max_size (inclusive), False otherwise
        """
        return 0 <= size <= max_size
