"""
Port utility functions for testing.
"""

import socket
import requests
from typing import Dict, Any


def check_websocket_server_port(port: int) -> bool:
    """
    Check if a WebSocket server is running on the specified port.
    
    Args:
        port: Port number to check
        
    Returns:
        True if server is running, False otherwise
    """
    try:
        # Try to connect to the port
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.settimeout(1)
        result = sock.connect_ex(('localhost', port))
        sock.close()
        return result == 0
    except Exception:
        return False


def check_http_server_port(port: int) -> bool:
    """
    Check if an HTTP server is running on the specified port.
    
    Args:
        port: Port number to check
        
    Returns:
        True if server is running, False otherwise
    """
    try:
        response = requests.get(f"http://localhost:{port}", timeout=1)
        return response.status_code < 500  # Any response means server is running
    except Exception:
        return False


def find_free_port() -> int:
    """
    Find a free port for testing.
    
    Returns:
        Available port number
    """
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        s.bind(('', 0))
        s.listen(1)
        port = s.getsockname()[1]
    return port


def create_test_health_server(recordings_path: str, snapshots_path: str) -> Any:
    """
    Create a test health server with a free port.
    
    Args:
        recordings_path: Path to recordings directory
        snapshots_path: Path to snapshots directory
        
    Returns:
        HealthServer instance configured for testing
    """
    from src.health_server import HealthServer
    
    # Find a free port for the test health server
    free_port = find_free_port()
    
    # Create health server with test-specific port
    health_server = HealthServer(
        host="127.0.0.1",  # Use localhost for tests
        port=free_port,
        recordings_path=recordings_path,
        snapshots_path=snapshots_path
    )
    
    return health_server
