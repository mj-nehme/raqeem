import re

MIN_PASSWORD_LEN = 8


class SecurityService:
    """Stub SecurityService for tests."""

    def validate_password_strength(self, password: str) -> bool:
        if len(password) < MIN_PASSWORD_LEN:
            return False
        has_upper = any(c.isupper() for c in password)
        has_lower = any(c.islower() for c in password)
        has_digit = any(c.isdigit() for c in password)
        has_symbol = any(not c.isalnum() for c in password)
        return has_upper and has_lower and has_digit and has_symbol

    def validate_email(self, email: str) -> bool:
        return re.match(r"^[^@\s]+@[^@\s]+\.[^@\s]+$", email) is not None

    def validate_jwt_token(self, token: str) -> bool:
        return bool(token)
