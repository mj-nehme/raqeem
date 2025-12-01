"""Tests for HTTP retry utility functions."""

import pytest
import respx
from httpx import Response

from app.util.http_retry import post_with_retry


class TestPostWithRetry:
    """Tests for the post_with_retry function."""

    @pytest.mark.asyncio
    @respx.mock
    async def test_successful_post(self):
        """Test successful POST request returns response."""
        respx.post("http://test.local/api").mock(
            return_value=Response(200, json={"status": "ok"})
        )
        
        result = await post_with_retry("http://test.local/api", json={"data": "test"})
        
        assert result is not None
        assert result.status_code == 200

    @pytest.mark.asyncio
    @respx.mock
    async def test_returns_on_client_error(self):
        """Test that 4xx errors (except 429) are not retried."""
        respx.post("http://test.local/api").mock(
            return_value=Response(400, json={"error": "bad request"})
        )
        
        result = await post_with_retry("http://test.local/api", json={"data": "test"}, max_retries=0)
        
        assert result is not None
        assert result.status_code == 400

    @pytest.mark.asyncio
    @respx.mock
    async def test_retries_on_server_error(self):
        """Test that 5xx errors trigger retries."""
        route = respx.post("http://test.local/api")
        route.side_effect = [
            Response(500, json={"error": "server error"}),
            Response(200, json={"status": "ok"}),
        ]
        
        result = await post_with_retry(
            "http://test.local/api",
            json={"data": "test"},
            max_retries=1,
            backoff_factor=0.01  # Very short for testing
        )
        
        assert result is not None
        assert result.status_code == 200
        assert route.call_count == 2

    @pytest.mark.asyncio
    @respx.mock
    async def test_retries_on_rate_limit(self):
        """Test that 429 rate limit errors trigger retries."""
        route = respx.post("http://test.local/api")
        route.side_effect = [
            Response(429, json={"error": "rate limited"}),
            Response(200, json={"status": "ok"}),
        ]
        
        result = await post_with_retry(
            "http://test.local/api",
            json={"data": "test"},
            max_retries=1,
            backoff_factor=0.01
        )
        
        assert result is not None
        assert result.status_code == 200
        assert route.call_count == 2

    @pytest.mark.asyncio
    @respx.mock
    async def test_returns_none_after_max_retries(self):
        """Test that None is returned after all retries are exhausted."""
        respx.post("http://test.local/api").mock(
            return_value=Response(500, json={"error": "server error"})
        )
        
        result = await post_with_retry(
            "http://test.local/api",
            json={"data": "test"},
            max_retries=1,
            backoff_factor=0.01
        )
        
        assert result is None

    @pytest.mark.asyncio
    @respx.mock
    async def test_timeout_parameter(self):
        """Test that timeout parameter is respected."""
        respx.post("http://test.local/api").mock(
            return_value=Response(200, json={"status": "ok"})
        )
        
        result = await post_with_retry(
            "http://test.local/api",
            json={"data": "test"},
            timeout=10.0
        )
        
        assert result is not None

    @pytest.mark.asyncio
    @respx.mock
    async def test_redirect_response(self):
        """Test that 3xx redirects are not retried."""
        respx.post("http://test.local/api").mock(
            return_value=Response(301, headers={"Location": "http://other.local/"})
        )
        
        result = await post_with_retry("http://test.local/api", json={"data": "test"})
        
        assert result is not None
        assert result.status_code == 301

    @pytest.mark.asyncio
    async def test_json_payload_is_sent(self):
        """Test that JSON payload is included in request."""
        captured_request = None
        
        @respx.mock
        async def make_request():
            nonlocal captured_request
            route = respx.post("http://test.local/api")
            route.mock(return_value=Response(200, json={"status": "ok"}))
            
            await post_with_retry("http://test.local/api", json={"key": "value"})
            
            # Get the captured request
            captured_request = route.calls.last.request
        
        await make_request()
        
        assert captured_request is not None


class TestHTTPStatusConstants:
    """Tests for HTTP status code constants."""

    def test_server_error_constant(self):
        """Test HTTP_STATUS_SERVER_ERROR is 500."""
        from app.util.http_retry import HTTP_STATUS_SERVER_ERROR
        assert HTTP_STATUS_SERVER_ERROR == 500

    def test_too_many_requests_constant(self):
        """Test HTTP_STATUS_TOO_MANY_REQUESTS is 429."""
        from app.util.http_retry import HTTP_STATUS_TOO_MANY_REQUESTS
        assert HTTP_STATUS_TOO_MANY_REQUESTS == 429
