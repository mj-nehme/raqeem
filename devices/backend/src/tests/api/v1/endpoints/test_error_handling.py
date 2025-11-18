"""
Focused error handling tests for devices endpoints.

Tests error validation and handling without requiring database connectivity.
Integration tests cover database-dependent scenarios.
"""

import uuid

import pytest
from app.main import app
from fastapi.testclient import TestClient

client = TestClient(app)


class TestDeviceRegistrationValidation:
    """Test input validation in device registration."""

    def test_register_device_rejects_legacy_id_field(self):
        """Test that legacy 'id' field is rejected with clear error."""
        response = client.post(
            "/api/v1/devices/register",
            json={
                "id": str(uuid.uuid4()),  # Legacy field
                "device_name": "test-device",
                "devicetype": "laptop",
            },
        )
        assert response.status_code == 400
        assert "unsupported legacy field: id" in response.json()["detail"]

    def test_register_device_rejects_legacy_name_field(self):
        """Test that legacy 'name' field is rejected."""
        response = client.post(
            "/api/v1/devices/register",
            json={
                "deviceid": str(uuid.uuid4()),
                "name": "test-device",  # Legacy field
                "devicetype": "laptop",
            },
        )
        assert response.status_code == 400
        assert "unsupported legacy field: name" in response.json()["detail"]

    def test_register_device_rejects_legacy_location_field(self):
        """Test that legacy 'location' field is rejected."""
        response = client.post(
            "/api/v1/devices/register",
            json={
                "deviceid": str(uuid.uuid4()),
                "device_name": "test-device",
                "location": "office",  # Legacy field
                "devicetype": "laptop",
            },
        )
        assert response.status_code == 400
        assert "unsupported legacy field: location" in response.json()["detail"]

    def test_register_device_requires_deviceid(self):
        """Test error when deviceid is missing."""
        response = client.post(
            "/api/v1/devices/register",
            json={
                "device_name": "test-device",
                "devicetype": "laptop",
            },
        )
        assert response.status_code == 400
        assert "missing required field: deviceid" in response.json()["detail"]

    def test_register_device_validates_deviceid_format(self):
        """Test error when deviceid is not a valid UUID."""
        response = client.post(
            "/api/v1/devices/register",
            json={
                "deviceid": "not-a-uuid",
                "device_name": "test-device",
                "devicetype": "laptop",
            },
        )
        assert response.status_code == 400
        assert "deviceid must be a valid UUID format" in response.json()["detail"]

    def test_register_device_rejects_empty_deviceid(self):
        """Test error when deviceid is empty string."""
        response = client.post(
            "/api/v1/devices/register",
            json={
                "deviceid": "",
                "device_name": "test-device",
                "devicetype": "laptop",
            },
        )
        assert response.status_code == 400

    def test_register_device_handles_malformed_json(self):
        """Test error handling with malformed JSON."""
        response = client.post(
            "/api/v1/devices/register",
            content="not-json",
            headers={"Content-Type": "application/json"},
        )
        assert response.status_code == 422  # FastAPI validation error


class TestInputSanitization:
    """Test input validation and boundary handling."""

    @pytest.mark.skip(reason="Requires database connection")
    def test_device_name_with_special_characters_accepted(self):
        """Test that special characters in device name are handled."""
        device_id = str(uuid.uuid4())
        # Just verify the endpoint accepts or rejects cleanly
        response = client.post(
            "/api/v1/devices/register",
            json={
                "deviceid": device_id,
                "device_name": "test<>&\"'device",
                "devicetype": "laptop",
            },
        )
        # Should handle gracefully (either accept or reject with clear error)
        assert response.status_code in [200, 201, 400, 422]

    @pytest.mark.skip(reason="Requires database connection")
    def test_extremely_long_device_name_handled(self):
        """Test handling of extremely long device names."""
        device_id = str(uuid.uuid4())
        response = client.post(
            "/api/v1/devices/register",
            json={
                "deviceid": device_id,
                "device_name": "x" * 10000,  # Very long name
                "devicetype": "laptop",
            },
        )
        # Should handle gracefully
        assert response.status_code in [200, 201, 400, 422]

    @pytest.mark.skip(reason="Requires database connection")
    def test_empty_device_name_validation(self):
        """Test validation of empty device name."""
        device_id = str(uuid.uuid4())
        response = client.post(
            "/api/v1/devices/register",
            json={
                "deviceid": device_id,
                "device_name": "",
                "devicetype": "laptop",
            },
        )
        # Validation may require non-empty name
        assert response.status_code in [200, 201, 400, 422]


class TestErrorResponseFormat:
    """Test that error responses follow consistent format."""

    def test_validation_error_includes_detail(self):
        """Test that validation errors include detail field."""
        response = client.post(
            "/api/v1/devices/register",
            json={
                "id": "legacy-field",  # Legacy field triggers validation
            },
        )
        assert response.status_code == 400
        error_data = response.json()
        assert "detail" in error_data
        assert isinstance(error_data["detail"], str)

    def test_uuid_validation_error_includes_context(self):
        """Test that UUID validation errors provide context."""
        response = client.post(
            "/api/v1/devices/register",
            json={
                "deviceid": "invalid-uuid-format",
                "device_name": "test",
                "devicetype": "laptop",
            },
        )
        assert response.status_code == 400
        error_detail = response.json()["detail"]
        assert "UUID" in error_detail or "uuid" in error_detail.lower()


class TestIdempotency:
    """Test idempotent operations."""

    @pytest.mark.skip(reason="Requires database connection")
    def test_register_same_device_twice_is_idempotent(self):
        """Test that registering same device twice is handled gracefully."""
        device_id = str(uuid.uuid4())
        device_data = {
            "deviceid": device_id,
            "device_name": "test-device",
            "devicetype": "laptop",
        }

        # First registration
        response1 = client.post("/api/v1/devices/register", json=device_data)

        # Second registration with same ID
        response2 = client.post("/api/v1/devices/register", json=device_data)

        # Both should succeed (idempotent operation)
        # May return different status codes but both should be successful
        assert response1.status_code in [200, 201]
        assert response2.status_code in [200, 201]


class TestAPIDocumentation:
    """Test that API documentation endpoints are available."""

    def test_openapi_schema_accessible(self):
        """Test that OpenAPI schema is accessible."""
        response = client.get("/openapi.json")
        assert response.status_code == 200
        schema = response.json()
        assert "openapi" in schema
        assert "paths" in schema

    def test_swagger_ui_accessible(self):
        """Test that Swagger UI is accessible."""
        response = client.get("/docs")
        assert response.status_code == 200

    def test_redoc_accessible(self):
        """Test that ReDoc is accessible."""
        response = client.get("/redoc")
        assert response.status_code == 200


class TestHealthEndpoint:
    """Test health check endpoint."""

    def test_health_endpoint_accessible(self):
        """Test that health endpoint is accessible."""
        # Try both possible paths
        response1 = client.get("/api/v1/health")
        response2 = client.get("/health")

        # At least one should work
        assert (response1.status_code == 200 or response2.status_code == 200)

        # Check the response that worked
        if response1.status_code == 200:
            data = response1.json()
            assert "status" in data
        elif response2.status_code == 200:
            data = response2.json()
            assert "status" in data

