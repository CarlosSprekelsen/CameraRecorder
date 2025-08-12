"""
Async mock helpers for MediaMTX controller tests.

Provides proper async context manager mocks for aiohttp session methods.
"""

from unittest.mock import Mock, AsyncMock
import asyncio


class AsyncContextManagerMock:
    """Mock async context manager for aiohttp session methods."""
    
    def __init__(self, response):
        self.response = response
    
    async def __aenter__(self):
        return self.response
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        pass


class MockResponse:
    """Mock HTTP response for testing."""
    
    def __init__(self, status=200, json_data=None, text_data=""):
        self.status = status
        self._json_data = json_data or {}
        self._text_data = text_data
    
    async def json(self):
        return self._json_data
    
    async def text(self):
        return self._text_data


def create_mock_session():
    """Create a properly mocked aiohttp session for testing."""
    mock_session = Mock()
    
    def mock_http_method(method_name):
        """Create a mock HTTP method that returns an async context manager."""
        def _mock_method(*args, **kwargs):
            # Default to success response
            response = MockResponse(200, {"status": "ok"})
            return AsyncContextManagerMock(response)
        return _mock_method
    
    # Mock the HTTP methods
    mock_session.get = mock_http_method("get")
    mock_session.post = mock_http_method("post")
    mock_session.put = mock_http_method("put")
    mock_session.delete = mock_http_method("delete")
    
    return mock_session


def create_mock_session_with_responses(responses):
    """Create a mock session that returns specific responses in sequence."""
    mock_session = Mock()
    response_index = [0]  # Use list to make it mutable in closure
    
    def mock_http_method(method_name):
        """Create a mock HTTP method that returns responses in sequence."""
        def _mock_method(*args, **kwargs):
            if response_index[0] < len(responses):
                response = responses[response_index[0]]
                response_index[0] += 1
                return AsyncContextManagerMock(response)
            else:
                # Default response if we run out of responses
                default_response = MockResponse(200, {"status": "ok"})
                return AsyncContextManagerMock(default_response)
        return _mock_method
    
    # Mock the HTTP methods
    mock_session.get = mock_http_method("get")
    mock_session.post = mock_http_method("post")
    mock_session.put = mock_http_method("put")
    mock_session.delete = mock_http_method("delete")
    
    return mock_session


def create_failure_response(status_code, error_text=""):
    """Create a mock response that simulates a failure."""
    return MockResponse(status_code, {"error": error_text}, error_text)


def create_success_response(json_data=None):
    """Create a mock response that simulates a success."""
    return MockResponse(200, json_data or {"status": "ok"})


def create_health_check_response(version="1.0.0", uptime=1200):
    """Create a mock response for health check endpoints."""
    return MockResponse(200, {
        "serverVersion": version,
        "serverUptime": uptime
    })


def mock_session_method(session, method_name, response):
    """Mock a specific session method to return a response wrapped in async context manager."""
    def _mock_method(*args, **kwargs):
        return AsyncContextManagerMock(response)
    setattr(session, method_name, _mock_method)


def mock_session_method_with_side_effect(session, method_name, side_effect):
    """Mock a session method with a side effect that returns async context managers."""
    def _mock_method(*args, **kwargs):
        result = side_effect(*args, **kwargs)
        if isinstance(result, MockResponse):
            return AsyncContextManagerMock(result)
        elif isinstance(result, Exception):
            raise result
        else:
            # Assume it's already an async context manager
            return result
    setattr(session, method_name, _mock_method)


def create_async_mock_with_response(response):
    """Create an AsyncMock that returns a response wrapped in async context manager."""
    async_mock = AsyncMock()
    async_mock.return_value = AsyncContextManagerMock(response)
    return async_mock


def create_async_mock_with_side_effect(side_effect):
    """Create an AsyncMock with side effect that handles responses properly."""
    async_mock = AsyncMock()
    
    def _side_effect_wrapper(*args, **kwargs):
        result = side_effect(*args, **kwargs)
        if isinstance(result, MockResponse):
            return AsyncContextManagerMock(result)
        elif isinstance(result, Exception):
            raise result
        else:
            return result
    
    async_mock.side_effect = _side_effect_wrapper
    return async_mock
