from datetime import UTC, datetime, timedelta
from typing import Any, cast

import jwt
from app.core.config import settings
from passlib.context import CryptContext

pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")


def hash_password(password: str) -> str:
    # passlib returns an untyped value; coerce to str for typing
    return cast("str", pwd_context.hash(password))


def verify_password(plain_password: str, hashed_password: str) -> bool:
    # passlib verify returns an untyped value; coerce to bool
    return cast("bool", pwd_context.verify(plain_password, hashed_password))


def create_access_token(data: dict[str, Any], expires_delta: timedelta | None = None) -> str:
    to_encode = data.copy()
    expire = datetime.now(UTC) + (expires_delta or timedelta(minutes=settings.access_token_expire_minutes))
    to_encode.update({"exp": expire})
    encoded_jwt = jwt.encode(to_encode, settings.secret_key, algorithm="HS256")
    return cast("str", encoded_jwt)


def decode_access_token(token: str) -> dict[str, Any]:
    try:
        payload = jwt.decode(token, settings.secret_key, algorithms=["HS256"])
        return cast("dict[str, Any]", payload)
    except jwt.PyJWTError:
        return {}
