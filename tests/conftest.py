import os
import pytest


@pytest.fixture
def devices_url():
    # Default to local devices backend; allow override via env
    return os.environ.get("DEVICES_URL", "http://localhost:8081")


@pytest.fixture
def mentor_url():
    # Default to local mentor backend; allow override via env
    return os.environ.get("MENTOR_URL", "http://localhost:8080")


@pytest.fixture(autouse=False)
def skip_smoke_if_disabled(request):
    # Skip the smoke test unless explicitly enabled (prevents CI failures when services aren't running)
    if request.node.name == "test_alert_flow":
        if os.environ.get("RUN_SMOKE_TESTS", "0") != "1":
            pytest.skip("Skipping smoke test; set RUN_SMOKE_TESTS=1 to enable.")
