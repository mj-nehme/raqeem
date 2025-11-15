import os
from fastapi.middleware.cors import CORSMiddleware

def setup_cors(app):
    """Configure CORS dynamically while keeping security sane.

    Priority:
    1. Use explicit origins from FRONTEND_ORIGINS (comma separated).
    2. Fallback to a curated localhost list (no wildcard) for dev when unset.
    3. If a wildcard is explicitly provided via env, disable credentials automatically.
    """
    raw = os.getenv("FRONTEND_ORIGINS", "")
    origins = [o.strip() for o in raw.split(",") if o.strip()]

    if len(origins) == 0:
        # Development defaults – keep list minimal and editable
        origins = [
            "http://localhost:4000",
            "http://localhost:5000",
        ]

    wildcard = len(origins) == 1 and origins[0] == "*"

    # Allow credentials only when we have explicit origins (not wildcard).
    allow_credentials = not wildcard and os.getenv("CORS_ALLOW_CREDENTIALS", "false").lower() in ["1", "true", "yes"]

    # Tighten headers/methods while remaining extensible via env overrides.
    methods_env = os.getenv("CORS_ALLOW_METHODS", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
    allow_methods = [m.strip().upper() for m in methods_env.split(",") if m.strip()]

    headers_env = os.getenv("CORS_ALLOW_HEADERS", "Authorization,Content-Type,Accept")
    allow_headers = [h.strip() for h in headers_env.split(",") if h.strip()]

    expose_env = os.getenv("CORS_EXPOSE_HEADERS", "Content-Length")
    expose_headers = [h.strip() for h in expose_env.split(",") if h.strip()]

    max_age = int(os.getenv("CORS_MAX_AGE", "600"))  # cache preflight for 10 minutes by default

    app.add_middleware(
        CORSMiddleware,
        allow_origins=origins,
        allow_credentials=allow_credentials,
        allow_methods=allow_methods,
        allow_headers=allow_headers,
        expose_headers=expose_headers,
        max_age=max_age,
    )
