class DeviceService:
    """Stub DeviceService for tests; methods are patched in tests."""

    def validate_device_data(self, _data):
        return True

    def is_device_online(self, _last_seen):
        return True

    def validate_device_type(self, t):
        return t in {"laptop", "desktop", "server", "mobile", "tablet"}
