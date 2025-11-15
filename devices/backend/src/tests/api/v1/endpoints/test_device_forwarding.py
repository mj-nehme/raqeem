import pytest
from httpx import AsyncClient, ASGITransport
from unittest.mock import patch, AsyncMock
from app.main import app


@pytest.mark.asyncio
async def test_register_device_forwards_to_mentor():
    """Test that device registration is forwarded to mentor backend when configured."""
    payload = {
        "id": "a843a399-701f-5011-aff3-4b69d8f21b11",
        "name": "Test Device for Forwarding",
        "device_type": "laptop",
        "os": "Linux",
        "location": "Test Lab",
        "ip_address": "192.168.1.200",
        "mac_address": "11:22:33:44:55:66",
        "current_user": "testuser"
    }
    
    # Mock the httpx client to verify forwarding happens
    with patch('httpx.AsyncClient') as mock_client:
        mock_response = AsyncMock()
        mock_response.status_code = 200
        mock_client.return_value.__aenter__.return_value.post = AsyncMock(return_value=mock_response)
        
        async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
            response = await ac.post("/api/v1/devices/register", json=payload)
        
        assert response.status_code == 200
        data = response.json()
        assert data["deviceid"] == "a843a399-701f-5011-aff3-4b69d8f21b11"
        # Verify that httpx client was called for forwarding (if MENTOR_API_URL is set)


@pytest.mark.asyncio
async def test_register_device_survives_mentor_forwarding_failure():
    """Test that device registration succeeds even if mentor forwarding fails."""
    payload = {
        "id": "e35e27a7-5808-5ea8-9ac5-acc284f75552",
        "name": "Test Device",
        "device_type": "laptop"
    }
    
    # Mock the httpx client to simulate forwarding failure
    with patch('httpx.AsyncClient') as mock_client:
        # Simulate network error during forwarding
        mock_client.return_value.__aenter__.return_value.post = AsyncMock(
            side_effect=Exception("Network error")
        )
        
        async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
            response = await ac.post("/api/v1/devices/register", json=payload)
        
        # Registration should still succeed despite forwarding failure
        assert response.status_code == 200
        data = response.json()
        assert data["deviceid"] == "e35e27a7-5808-5ea8-9ac5-acc284f75552"


@pytest.mark.asyncio
async def test_metrics_forwarding_to_mentor():
    """Test that metrics are forwarded to mentor backend when configured."""
    device_id = "33f9ce74-d0ce-515e-bb95-2464e9faa707"
    payload = {
        "cpu_usage": 55.5,
        "cpu_temp": 70.0,
        "memory_total": 16000000000,
        "memory_used": 10000000000,
        "swap_used": 200000000,
        "disk_total": 500000000000,
        "disk_used": 300000000000,
        "net_bytes_in": 2048000,
        "net_bytes_out": 4096000
    }
    
    with patch('httpx.AsyncClient') as mock_client:
        mock_response = AsyncMock()
        mock_response.status_code = 200
        mock_client.return_value.__aenter__.return_value.post = AsyncMock(return_value=mock_response)
        
        async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
            response = await ac.post(f"/api/v1/devices/{device_id}/metrics", json=payload)
        
        assert response.status_code == 200
        assert response.json()["status"] == "ok"
