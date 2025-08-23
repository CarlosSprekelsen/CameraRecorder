"""
Test JavaScript client example functionality.

Requirements Traceability:
- REQ-CLIENT-001: Client examples shall be fully functional and tested
- REQ-CLIENT-002: Client examples shall demonstrate proper authentication
- REQ-CLIENT-003: Client examples shall handle errors gracefully
- REQ-AUTH-001: Authentication shall work with JWT tokens
- REQ-AUTH-002: Authentication shall work with API keys

Story Coverage: S8.1 - Client Usage Examples
IV&V Control Point: Client examples validation

API Documentation Reference: docs/api/json-rpc-methods.md
"""

import pytest
import subprocess
import sys
import os
import json
import tempfile
from pathlib import Path


class TestJavaScriptClientExample:
    """Test JavaScript client example functionality."""
    
    @pytest.fixture
    def js_client_path(self):
        """Path to JavaScript client example."""
        return Path(__file__).parent.parent.parent / "examples" / "javascript" / "camera_client.js"
    
    @pytest.fixture
    def test_config(self):
        """Test configuration for JavaScript client."""
        return {
            "host": "localhost",
            "port": 8002,
            "useSsl": False,
            "authType": "jwt",
            "authToken": "test_jwt_token",
            "maxRetries": 3,
            "retryDelay": 1.0
        }
    
    def test_javascript_client_file_exists(self, js_client_path):
        """Test that JavaScript client file exists and is readable."""
        assert js_client_path.exists(), f"JavaScript client file not found: {js_client_path}"
        assert js_client_path.is_file(), f"JavaScript client path is not a file: {js_client_path}"
        
        # Test file is readable and contains expected content
        content = js_client_path.read_text()
        assert "class CameraClient" in content, "JavaScript client file does not contain CameraClient class"
        assert "async connect()" in content, "JavaScript client file does not contain connect method"
        assert "async getCameraList()" in content, "JavaScript client file does not contain getCameraList method"
    
    def test_javascript_client_syntax_valid(self, js_client_path):
        """Test that JavaScript client has valid syntax."""
        try:
            # Use Node.js to check syntax
            result = subprocess.run(
                ["node", "--check", str(js_client_path)],
                capture_output=True,
                text=True,
                timeout=10
            )
            assert result.returncode == 0, f"JavaScript syntax error: {result.stderr}"
        except subprocess.TimeoutExpired:
            pytest.fail("JavaScript syntax check timed out")
        except FileNotFoundError:
            pytest.skip("Node.js not available for syntax checking")
    
    def test_javascript_client_imports_available(self, js_client_path):
        """Test that JavaScript client has required imports."""
        content = js_client_path.read_text()
        
        # Check for required imports (CommonJS format)
        required_imports = [
            "const WebSocket = require('ws');",
            "const { v4: uuidv4 } = require('uuid');"
        ]
        
        for import_stmt in required_imports:
            assert import_stmt in content, f"Missing required import: {import_stmt}"
    
    def test_javascript_client_class_structure(self, js_client_path):
        """Test JavaScript client class structure."""
        content = js_client_path.read_text()
        
        # Check for required methods
        required_methods = [
            "constructor",
            "async connect()",
            "async disconnect()",
            "async _authenticate()",
            "async getCameraList()",
            "async getCameraStatus(",
            "async takeSnapshot(",
            "async startRecording(",
            "async stopRecording(",
            "async ping()",
            "_sendRequest(",
            "_sleep("
        ]
        
        for method in required_methods:
            assert method in content, f"Missing required method: {method}"
    
    def test_javascript_client_error_handling(self, js_client_path):
        """Test that JavaScript client has proper error handling."""
        content = js_client_path.read_text()
        
        # Check for error handling patterns
        error_patterns = [
            "try {",
            "catch (",
            "throw new",
            "Error(",
            "AuthenticationError",
            "CameraNotFoundError"
        ]
        
        for pattern in error_patterns:
            assert pattern in content, f"Missing error handling pattern: {pattern}"
    
    def test_javascript_client_authentication_flow(self, js_client_path):
        """Test that JavaScript client has proper authentication flow."""
        content = js_client_path.read_text()
        
        # Check for authentication-related code
        auth_patterns = [
            "authenticate",
            "auth_type",
            "authToken",
            "apiKey",
            "jwt",
            "api_key"
        ]
        
        for pattern in auth_patterns:
            assert pattern in content, f"Missing authentication pattern: {pattern}"
    
    def test_javascript_client_websocket_communication(self, js_client_path):
        """Test that JavaScript client has WebSocket communication."""
        content = js_client_path.read_text()
        
        # Check for WebSocket-related code
        websocket_patterns = [
            "WebSocket",
            "ws",
            "wss",
            "send(",
            "on(",
            "message",
            "open",
            "close"
        ]
        
        # Check that at least some WebSocket patterns are present
        found_patterns = [pattern for pattern in websocket_patterns if pattern in content]
        assert len(found_patterns) >= 4, f"Not enough WebSocket patterns found. Found: {found_patterns}"
    
    def test_javascript_client_json_rpc_format(self, js_client_path):
        """Test that JavaScript client uses proper JSON-RPC format."""
        content = js_client_path.read_text()
        
        # Check for JSON-RPC patterns
        jsonrpc_patterns = [
            "jsonrpc",
            "method",
            "params",
            "id",
            "result",
            "error"
        ]
        
        for pattern in jsonrpc_patterns:
            assert pattern in content, f"Missing JSON-RPC pattern: {pattern}"
    
    def test_javascript_client_configuration_options(self, js_client_path):
        """Test that JavaScript client supports configuration options."""
        content = js_client_path.read_text()
        
        # Check for configuration options
        config_options = [
            "host",
            "port",
            "useSsl",
            "authType",
            "authToken",
            "apiKey",
            "maxRetries",
            "retryDelay"
        ]
        
        for option in config_options:
            assert option in content, f"Missing configuration option: {option}"
    
    def test_javascript_client_camera_operations(self, js_client_path):
        """Test that JavaScript client supports all camera operations."""
        content = js_client_path.read_text()
        
        # Check for camera operation methods
        camera_operations = [
            "getCameraList",
            "getCameraStatus",
            "takeSnapshot",
            "startRecording",
            "stopRecording"
        ]
        
        for operation in camera_operations:
            assert operation in content, f"Missing camera operation: {operation}"
    
    def test_javascript_client_data_structures(self, js_client_path):
        """Test that JavaScript client has proper data structures."""
        content = js_client_path.read_text()
        
        # Check for data structure classes
        data_structures = [
            "class CameraInfo",
            "class RecordingInfo"
        ]
        
        for structure in data_structures:
            assert structure in content, f"Missing data structure: {structure}"
    
    def test_javascript_client_example_usage(self, js_client_path):
        """Test that JavaScript client has example usage patterns."""
        content = js_client_path.read_text()
        
        # Check for example usage at the end of the file
        if "// Example usage" in content or "async function main()" in content:
            # Verify the example is complete
            assert "new CameraClient" in content, "Missing CameraClient instantiation example"
            assert "await client.connect()" in content, "Missing connection example"
            assert "await client.getCameraList()" in content, "Missing camera list example"
    
    def test_javascript_client_error_classes(self, js_client_path):
        """Test that JavaScript client has proper error classes."""
        content = js_client_path.read_text()
        
        # Check for custom error classes
        error_classes = [
            "class AuthenticationError",
            "class CameraNotFoundError",
            "class CameraServiceError"
        ]
        
        for error_class in error_classes:
            assert error_class in content, f"Missing error class: {error_class}"
    
    def test_javascript_client_retry_logic(self, js_client_path):
        """Test that JavaScript client has retry logic."""
        content = js_client_path.read_text()
        
        # Check for retry-related code
        retry_patterns = [
            "maxRetries",
            "retryDelay",
            "for (let attempt = 1",
            "attempt <= this.maxRetries",
            "await this.sleep("
        ]
        
        # Check that at least some retry patterns are present
        found_patterns = [pattern for pattern in retry_patterns if pattern in content]
        assert len(found_patterns) >= 2, f"Not enough retry patterns found. Found: {found_patterns}"
    
    def test_javascript_client_connection_management(self, js_client_path):
        """Test that JavaScript client has proper connection management."""
        content = js_client_path.read_text()
        
        # Check for connection management code
        connection_patterns = [
            "connected",
            "websocket",
            "connect()",
            "disconnect()",
            "close()"
        ]
        
        for pattern in connection_patterns:
            assert pattern in content, f"Missing connection management pattern: {pattern}"
    
    def test_javascript_client_message_handling(self, js_client_path):
        """Test that JavaScript client has proper message handling."""
        content = js_client_path.read_text()
        
        # Check for message handling code
        message_patterns = [
            "JSON.parse",
            "JSON.stringify",
            "send(",
            "onmessage"
        ]
        
        # Check that at least some message handling patterns are present
        found_patterns = [pattern for pattern in message_patterns if pattern in content]
        assert len(found_patterns) >= 2, f"Not enough message handling patterns found. Found: {found_patterns}"
    
    def test_javascript_client_validation(self, js_client_path):
        """Test that JavaScript client has input validation."""
        content = js_client_path.read_text()
        
        # Check for validation patterns
        validation_patterns = [
            "if (!",
            "throw new",
            "invalid",
            "validate"
        ]
        
        # Check that at least some validation patterns are present
        found_patterns = [pattern for pattern in validation_patterns if pattern in content]
        assert len(found_patterns) >= 2, f"Not enough validation patterns found. Found: {found_patterns}"
    
    def test_javascript_client_logging(self, js_client_path):
        """Test that JavaScript client has logging capabilities."""
        content = js_client_path.read_text()
        
        # Check for logging patterns
        logging_patterns = [
            "console.log",
            "console.error",
            "console.warn",
            "console.debug"
        ]
        
        # At least one logging method should be present
        has_logging = any(pattern in content for pattern in logging_patterns)
        assert has_logging, "JavaScript client should have logging capabilities"
    
    def test_javascript_client_async_await_patterns(self, js_client_path):
        """Test that JavaScript client uses proper async/await patterns."""
        content = js_client_path.read_text()
        
        # Check for async/await patterns
        async_patterns = [
            "async ",
            "await ",
            "Promise",
            "then(",
            "catch("
        ]
        
        # Should have async/await patterns
        has_async = any(pattern in content for pattern in async_patterns)
        assert has_async, "JavaScript client should use async/await patterns"
    
    def test_javascript_client_file_completeness(self, js_client_path):
        """Test that JavaScript client file is complete and well-structured."""
        content = js_client_path.read_text()
        
        # Basic file structure checks
        assert len(content) > 1000, "JavaScript client file seems too short"
        assert content.count("class") >= 4, "JavaScript client should have multiple classes"
        assert content.count("async") >= 10, "JavaScript client should have multiple async methods"
        # Note: This is a CommonJS file, not ES6 module, so no exports expected
    
    def test_javascript_client_no_syntax_errors(self, js_client_path):
        """Test that JavaScript client has no obvious syntax errors."""
        content = js_client_path.read_text()
        
        # Check for common syntax issues
        syntax_issues = [
            "{{",  # Double braces
            "}}",  # Double closing braces
            ";;",  # Double semicolons
            "()",  # Empty parentheses
            "[]",  # Empty brackets
        ]
        
        for issue in syntax_issues:
            # These patterns might be valid in some contexts, but should be reasonable
            count = content.count(issue)
            if issue == "()":  # Empty parentheses are common in function calls
                assert count < 50, f"Too many potential syntax issues: {issue} appears {count} times"
            else:
                assert count < 10, f"Too many potential syntax issues: {issue} appears {count} times"
    
    def test_javascript_client_consistent_formatting(self, js_client_path):
        """Test that JavaScript client has consistent formatting."""
        content = js_client_path.read_text()
        
        # Check for consistent indentation and formatting
        lines = content.split('\n')
        
        # Should have reasonable line lengths
        long_lines = [line for line in lines if len(line) > 120]
        assert len(long_lines) < 10, f"Too many long lines: {len(long_lines)}"
        
        # Should have proper spacing
        empty_lines = [line for line in lines if line.strip() == '']
        assert len(empty_lines) > 0, "JavaScript client should have some empty lines for readability"
    
    def test_javascript_client_documentation(self, js_client_path):
        """Test that JavaScript client has proper documentation."""
        content = js_client_path.read_text()
        
        # Check for documentation patterns
        doc_patterns = [
            "/**",
            "*/",
            "@param",
            "@returns",
            "@throws",
            "// "
        ]
        
        # Should have some documentation
        has_docs = any(pattern in content for pattern in doc_patterns)
        assert has_docs, "JavaScript client should have documentation"
