"""Real tests for SecurityService - no mocking, direct function calls."""

import pytest
from app.services.security_service import SecurityService


class TestSecurityServiceReal:
    """Test SecurityService with real function calls."""

    @pytest.fixture
    def service(self):
        """Create a real SecurityService instance."""
        return SecurityService()

    # Password Strength Validation Tests
    def test_validate_password_strong(self, service):
        """Test strong password is valid."""
        assert service.validate_password_strength("StrongPassword123!") is True

    def test_validate_password_with_special_chars(self, service):
        """Test password with various special characters is valid."""
        assert service.validate_password_strength("MySecure2024#Pass") is True

    def test_validate_password_complex(self, service):
        """Test complex password is valid."""
        assert service.validate_password_strength("Complex!Password@456") is True

    def test_validate_password_minimum_length(self, service):
        """Test password at exactly minimum length."""
        # 8 chars with upper, lower, digit, special
        assert service.validate_password_strength("Aa1!Bb2@") is True

    def test_validate_password_too_short(self, service):
        """Test password too short is invalid."""
        assert service.validate_password_strength("Aa1!") is False

    def test_validate_password_no_uppercase(self, service):
        """Test password without uppercase is invalid."""
        assert service.validate_password_strength("password123!") is False

    def test_validate_password_no_lowercase(self, service):
        """Test password without lowercase is invalid."""
        assert service.validate_password_strength("PASSWORD123!") is False

    def test_validate_password_no_digit(self, service):
        """Test password without digit is invalid."""
        assert service.validate_password_strength("PasswordOnly!") is False

    def test_validate_password_no_special(self, service):
        """Test password without special character is invalid."""
        assert service.validate_password_strength("Password123") is False

    def test_validate_password_empty(self, service):
        """Test empty password is invalid."""
        assert service.validate_password_strength("") is False

    def test_validate_password_only_numbers(self, service):
        """Test password with only numbers is invalid."""
        assert service.validate_password_strength("12345678") is False

    def test_validate_password_only_letters(self, service):
        """Test password with only letters is invalid."""
        assert service.validate_password_strength("abcdefgh") is False

    def test_validate_password_with_spaces(self, service):
        """Test password with spaces as special chars."""
        # Space counts as a non-alphanumeric character
        assert service.validate_password_strength("Pass word1") is True

    # Email Validation Tests
    def test_validate_email_valid_simple(self, service):
        """Test simple valid email."""
        assert service.validate_email("user@example.com") is True

    def test_validate_email_valid_subdomain(self, service):
        """Test email with subdomain is valid."""
        assert service.validate_email("user@subdomain.example.com") is True

    def test_validate_email_valid_with_dot(self, service):
        """Test email with dot in username is valid."""
        assert service.validate_email("test.email@domain.org") is True

    def test_validate_email_valid_with_plus(self, service):
        """Test email with plus sign is valid."""
        assert service.validate_email("user+tag@example.com") is True

    def test_validate_email_invalid_no_at(self, service):
        """Test email without @ is invalid."""
        assert service.validate_email("invalid-email") is False

    def test_validate_email_invalid_no_domain(self, service):
        """Test email without domain is invalid."""
        assert service.validate_email("user@") is False

    def test_validate_email_invalid_no_user(self, service):
        """Test email without username is invalid."""
        assert service.validate_email("@example.com") is False

    def test_validate_email_invalid_no_tld(self, service):
        """Test email without TLD is invalid."""
        assert service.validate_email("user@domain") is False

    def test_validate_email_invalid_double_at(self, service):
        """Test email with double @ is invalid."""
        assert service.validate_email("user@@example.com") is False

    def test_validate_email_empty(self, service):
        """Test empty email is invalid."""
        assert service.validate_email("") is False

    def test_validate_email_with_spaces(self, service):
        """Test email with spaces is invalid."""
        assert service.validate_email("user @example.com") is False

    # JWT Token Validation Tests
    def test_validate_jwt_token_valid(self, service):
        """Test valid JWT-like token."""
        assert service.validate_jwt_token("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U") is True

    def test_validate_jwt_token_simple(self, service):
        """Test simple non-empty token."""
        assert service.validate_jwt_token("some-token") is True

    def test_validate_jwt_token_empty(self, service):
        """Test empty token is invalid."""
        assert service.validate_jwt_token("") is False

    def test_validate_jwt_token_whitespace_only(self, service):
        """Test whitespace-only token is technically valid (non-empty)."""
        # Based on implementation, bool("   ") is True
        assert service.validate_jwt_token("   ") is True

    def test_validate_jwt_token_single_char(self, service):
        """Test single character token is valid."""
        assert service.validate_jwt_token("x") is True
