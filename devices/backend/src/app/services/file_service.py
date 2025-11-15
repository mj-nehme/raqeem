class FileService:
    """Stub FileService for tests."""
    IMAGE_TYPES = {"image/png","image/jpeg","image/gif"}
    def validate_file_type(self, filename: str, content_type: str) -> bool:
        return content_type in self.IMAGE_TYPES and any(filename.lower().endswith(ext) for ext in [".png",".jpg",".jpeg",".gif"])
    def validate_file_size(self, size: int, max_size: int) -> bool:
        return 0 <= size <= max_size
