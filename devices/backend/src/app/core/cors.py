import os
import re
from fastapi.middleware.cors import CORSMiddleware

def setup_cors(app):
    """Configure CORS dynamically while keeping security sane.

    Priority:
    1. Use explicit origins from FRONTEND_ORIGINS (comma separated).
    2. Fallback to a curated localhost list (no wildcard) for dev when unset.
    3. If a wildcard is explicitly provided via env, disable credentials automatically.
    """
    # FRONTEND_ORIGINS: explicit comma-separated list (highest precedence)
    raw = os.getenv("FRONTEND_ORIGINS", "")
    # FRONTEND_ORIGIN_REGEX: pattern for origins (e.g. ^http://localhost:4\\d{3}$)
    origin_regex_env = os.getenv("FRONTEND_ORIGIN_REGEX", "")
    origins = [o.strip() for o in raw.split(",") if o.strip()]

    allow_origin_regex = None
    if origin_regex_env:
        try:
            re.compile(origin_regex_env)
            allow_origin_regex = origin_regex_env
        except re.error:
            # Invalid regex ignored, will fallback to explicit origins
            pass

    if len(origins) == 0 and not allow_origin_regex:
        # Derive from dynamic port exported by startup scripts (start-smart.sh).
        # The devices frontend port is stored in VITE_DEVICES_FRONTEND_PORT at runtime, but we prefer
        # an explicit DEVICES_FRONTEND_PORT (set by start script) if available; fall back to 4000 base.
        dev_port = os.getenv("DEVICES_FRONTEND_PORT") or os.getenv("VITE_DEVICES_FRONTEND_PORT") or "4001"
        mentor_port = os.getenv("MENTOR_FRONTEND_PORT") or os.getenv("VITE_MENTOR_FRONTEND_PORT")
        # Always include devices frontend; optionally mentor frontend if defined.
        origins = [f"http://localhost:{dev_port}"]
        if mentor_port and mentor_port != dev_port:
            origins.append(f"http://localhost:{mentor_port}")

    wildcard = (len(origins) == 1 and origins[0] == "*") or (allow_origin_regex == ".*")

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

    cors_kwargs = dict(
        allow_credentials=allow_credentials,
        allow_methods=allow_methods,
        allow_headers=allow_headers,
        expose_headers=expose_headers,
        max_age=max_age,
    )
    if allow_origin_regex:
        cors_kwargs["allow_origin_regex"] = allow_origin_regex
    else:
        cors_kwargs["allow_origins"] = origins

    app.add_middleware(CORSMiddleware, **cors_kwargs)
