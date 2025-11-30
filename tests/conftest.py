import os
import pytest
import requests


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


def pytest_collection_modifyitems(config, items):
    """Skip integration tests if services are not running."""
    # Only skip if RUN_INTEGRATION_TESTS env var is explicitly set to 0
    if os.environ.get("RUN_INTEGRATION_TESTS", "1") == "0":
        skip_integration = pytest.mark.skip(reason="Integration tests disabled via RUN_INTEGRATION_TESTS=0")
        for item in items:
            if "integration" in str(item.fspath):
                item.add_marker(skip_integration)
        return
    
    # Check if services are running for integration tests
    devices_running = False
    mentor_running = False
    
    try:
        response = requests.get("http://localhost:8081/health", timeout=1)
        devices_running = response.status_code == 200
    except Exception:
        pass
    
    try:
        response = requests.get("http://localhost:8080/health", timeout=1)
        mentor_running = response.status_code == 200
    except Exception:
        pass
    
    # Only skip tests that need services if they're not running
    if not (devices_running and mentor_running):
        skip_no_services = pytest.mark.skip(
            reason="Services not running. Start services with ./start.sh or docker-compose up"
        )
        for item in items:
            # Skip tests in integration folder that are not observability tests (which have their own wait logic)
            if "integration" in str(item.fspath) and "observability" not in str(item.fspath):
                item.add_marker(skip_no_services)
