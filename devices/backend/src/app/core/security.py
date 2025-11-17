from datetime import datetime, timedelta, timezone
from typing import Any, cast

import jwt
from app.core.config import settings
from passlib.context import CryptContext

# Enhanced password hashing with stronger bcrypt rounds
# Default rounds=12 provides good security/performance balance
# Increase to 14+ for high-security applications
pwd_context = CryptContext(
    schemes=["bcrypt"],
    deprecated="auto",
    bcrypt__rounds=12,  # Explicit rounds for clarity and security
)


def hash_password(password: str) -> str:
    # passlib returns an untyped value; coerce to str for typing
    return cast("str", pwd_context.hash(password))


def verify_password(plain_password: str, hashed_password: str) -> bool:
    # passlib verify returns an untyped value; coerce to bool
    return cast("bool", pwd_context.verify(plain_password, hashed_password))


def create_access_token(data: dict[str, Any], expires_delta: timedelta | None = None) -> str:
    """Create a JWT access token with expiration.

    Args:
        data: Dictionary of claims to encode in the token.
        expires_delta: Optional custom expiration time. Defaults to settings value.

    Returns:
        Encoded JWT token string.

    Security Notes:
        - Uses HS256 algorithm (HMAC-SHA256)
        - Includes expiration claim to prevent token reuse
        - Tokens should be transmitted over HTTPS only
        - Consider implementing token refresh for long-lived sessions
    """
    to_encode = data.copy()
    expire = datetime.now(timezone.utc) + (expires_delta or timedelta(minutes=settings.access_token_expire_minutes))
    to_encode.update({"exp": expire})
    encoded_jwt = jwt.encode(to_encode, settings.secret_key, algorithm="HS256")
    return cast("str", encoded_jwt)


def decode_access_token(token: str) -> dict[str, Any]:
    """Decode and validate a JWT access token.

    Args:
        token: JWT token string to decode.

    Returns:
        Dictionary of decoded claims, or empty dict if invalid.

    Security Notes:
        - Validates token signature using secret key
        - Checks expiration automatically
        - Returns empty dict on any validation failure
        - Caller should always check returned dict is not empty
    """
    try:
        payload = jwt.decode(token, settings.secret_key, algorithms=["HS256"])
        return cast("dict[str, Any]", payload)
    except jwt.ExpiredSignatureError:
        # Token has expired
        return {}
    except jwt.InvalidTokenError:
        # Token is invalid (bad signature, malformed, etc.)
        return {}
    except Exception:
        # Catch-all for unexpected errors
        return {}
