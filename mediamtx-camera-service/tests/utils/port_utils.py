"""
Port utility functions for testing.

Provides utilities to check if ports are already in use for test coordination.
"""

import socket


def is_port_listening(host: str, port: int, timeout: float = 1.0) -> bool:
    """
    Check if a port is already listening on the specified host.
    
    Args:
        host: Host address to check
        port: Port number to check
        timeout: Socket timeout in seconds
        
    Returns:
        True if port is listening, False otherwise
    """
    try:
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.settimeout(timeout)
        result = sock.connect_ex((host, port))
        sock.close()
        return result == 0
    except Exception:
        return False


def check_websocket_server_port(port: int = 8002, host: str = "127.0.0.1") -> bool:
    """
    Check if the WebSocket server port is already in use.
    
    Args:
        port: Port to check (default 8002 for production)
        host: Host to check (default 127.0.0.1)
        
    Returns:
        True if port is in use, False otherwise
    """
    return is_port_listening(host, port)
