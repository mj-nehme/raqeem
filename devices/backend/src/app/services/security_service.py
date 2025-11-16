"""Security service for validating passwords, emails, and tokens."""

import re

MIN_PASSWORD_LEN = 8


class SecurityService:
    """Service for security-related validations.

    This is a stub service primarily used for testing.
    Methods can be patched in tests to customize behavior.
    """

    def validate_password_strength(self, password: str) -> bool:
        """Validate that a password meets security requirements.

        A valid password must:
        - Be at least 8 characters long
        - Contain at least one uppercase letter
        - Contain at least one lowercase letter
        - Contain at least one digit
        - Contain at least one special character

        Args:
            password: The password string to validate

        Returns:
            True if password meets all requirements, False otherwise
        """
        if len(password) < MIN_PASSWORD_LEN:
            return False
        has_upper = any(c.isupper() for c in password)
        has_lower = any(c.islower() for c in password)
        has_digit = any(c.isdigit() for c in password)
        has_symbol = any(not c.isalnum() for c in password)
        return has_upper and has_lower and has_digit and has_symbol

    def validate_email(self, email: str) -> bool:
        """Validate that an email address has a valid format.

        Args:
            email: The email address string to validate

        Returns:
            True if email format is valid, False otherwise

        Note:
            Uses a basic regex pattern for validation.
            Does not verify that the email address actually exists.
        """
        return re.match(r"^[^@\s]+@[^@\s]+\.[^@\s]+$", email) is not None

    def validate_jwt_token(self, token: str) -> bool:
        """Validate that a JWT token is present and non-empty.

        Args:
            token: The JWT token string to validate

        Returns:
            True if token is non-empty, False otherwise

        Note:
            This is a basic validation that only checks for presence.
            In production, this should verify signature and expiration.
        """
        return bool(token)
