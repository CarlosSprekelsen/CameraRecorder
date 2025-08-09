import asyncio
import json
import os
import pytest

from src.security.jwt_handler import JWTHandler
from src.websocket_server.server import WebSocketJsonRpcServer


@pytest.mark.asyncio
async def test_rejects_invalid_token(monkeypatch):
    # Arrange
    host = "127.0.0.1"
    port = 8765
    server = WebSocketJsonRpcServer(host=host, port=port, websocket_path="/ws", max_connections=10)

    # Inject security middleware via service manager startup path is heavy; patch minimal to accept but require auth
    class DummySecurity:
        def __init__(self):
            self._auth = {}
        def can_accept_connection(self, client_id):
            return True
        def register_connection(self, client_id):
            pass
        def unregister_connection(self, client_id):
            pass
        def check_rate_limit(self, client_id):
            return True
        def is_authenticated(self, client_id):
            return client_id in self._auth
        async def authenticate_connection(self, client_id, token, auth_type="auto"):
            class AR:
                authenticated = False
                error_message = "Invalid or expired JWT token"
                role = None
                user_id = None
                auth_method = None
                expires_at = None
            return AR()
        def has_permission(self, client_id, required_role):
            return False
        def get_auth_result(self, client_id):
            return None

    server.set_security_middleware(DummySecurity())
    await server.start()

    import websockets
    uri = f"ws://{host}:{port}/ws"
    async with websockets.connect(uri) as ws:
        # Attempt protected method with invalid token
        req = {"jsonrpc": "2.0", "id": 1, "method": "take_snapshot", "params": {"device": "/dev/video0", "auth_token": "bad.token"}}
        await ws.send(json.dumps(req))
        resp = json.loads(await ws.recv())
        print("INVALID_TOKEN_RESPONSE:", resp)
        assert resp["error"]["code"] == -32001
        assert "Authentication" in resp["error"]["message"]

    await server.stop()


@pytest.mark.asyncio
async def test_accepts_valid_token_and_enforces_role(monkeypatch):
    host = "127.0.0.1"
    port = 8766
    server = WebSocketJsonRpcServer(host=host, port=port, websocket_path="/ws", max_connections=10)

    secret = os.environ.get("CAMERA_SERVICE_JWT_SECRET", "test-secret-key")
    jwt_handler = JWTHandler(secret_key=secret)
    token_operator = jwt_handler.generate_token("user1", "operator", expiry_hours=1)

    class DummySecurity:
        def __init__(self):
            self._auth = {}
        def can_accept_connection(self, client_id):
            return True
        def register_connection(self, client_id):
            pass
        def unregister_connection(self, client_id):
            pass
        def check_rate_limit(self, client_id):
            return True
        def is_authenticated(self, client_id):
            return client_id in self._auth
        async def authenticate_connection(self, client_id, token, auth_type="auto"):
            class AR:
                pass
            ar = AR()
            if token == token_operator:
                ar.authenticated = True
                ar.error_message = None
                ar.role = "operator"
                ar.user_id = "user1"
                ar.auth_method = "jwt"
                ar.expires_at = 9999999999
                self._auth[client_id] = ar
            else:
                ar.authenticated = False
                ar.error_message = "Invalid token"
                ar.role = None
                ar.user_id = None
                ar.auth_method = None
                ar.expires_at = None
            return ar
        def has_permission(self, client_id, required_role):
            ar = self._auth.get(client_id)
            return ar and ar.role in ("operator", "admin")
        def get_auth_result(self, client_id):
            return self._auth.get(client_id)

    # Stub snapshot method to succeed
    async def fake_snapshot(params=None):
        return {"status": "SUCCESS", "device": params.get("device")}

    server.set_security_middleware(DummySecurity())
    await server.start()
    # Register after start to avoid overwrite by built-ins
    server.register_method("take_snapshot", fake_snapshot)

    import websockets
    uri = f"ws://{host}:{port}/ws"
    async with websockets.connect(uri) as ws:
        # Authenticate
        await ws.send(json.dumps({"jsonrpc": "2.0", "id": 1, "method": "authenticate", "params": {"token": token_operator}}))
        auth_resp = json.loads(await ws.recv())
        print("AUTH_SUCCESS_RESPONSE:", auth_resp)
        assert auth_resp["result"]["authenticated"] is True
        assert auth_resp["result"]["role"] == "operator"

        # Call protected method
        await ws.send(json.dumps({"jsonrpc": "2.0", "id": 2, "method": "take_snapshot", "params": {"device": "/dev/video0"}}))
        resp = json.loads(await ws.recv())
        print("PROTECTED_METHOD_RESPONSE:", resp)
        assert resp["result"]["status"] == "SUCCESS"

    await server.stop()


@pytest.mark.asyncio
async def test_expired_token_rejected(monkeypatch):
    host = "127.0.0.1"
    port = 8767
    server = WebSocketJsonRpcServer(host=host, port=port, websocket_path="/ws", max_connections=10)

    class DummySecurity:
        def __init__(self):
            self._auth = {}
        def can_accept_connection(self, client_id):
            return True
        def register_connection(self, client_id):
            pass
        def unregister_connection(self, client_id):
            pass
        def check_rate_limit(self, client_id):
            return True
        def is_authenticated(self, client_id):
            return client_id in self._auth
        async def authenticate_connection(self, client_id, token, auth_type="auto"):
            class AR:
                authenticated = True
                error_message = None
                role = "operator"
                user_id = "u"
                auth_method = "jwt"
                expires_at = 0  # expired
            self._auth[client_id] = AR()
            return self._auth[client_id]
        def has_permission(self, client_id, required_role):
            return True
        def get_auth_result(self, client_id):
            return self._auth.get(client_id)

    async def fake_snapshot(params=None):
        return {"status": "SUCCESS"}

    server.set_security_middleware(DummySecurity())
    await server.start()
    # Register after start to avoid overwrite by built-ins
    server.register_method("take_snapshot", fake_snapshot)

    import websockets
    uri = f"ws://{host}:{port}/ws"
    async with websockets.connect(uri) as ws:
        # Authenticate (expired)
        await ws.send(json.dumps({"jsonrpc": "2.0", "id": 1, "method": "authenticate", "params": {"token": "any"}}))
        _ = json.loads(await ws.recv())
        # Protected method must be rejected due to expiry
        await ws.send(json.dumps({"jsonrpc": "2.0", "id": 2, "method": "take_snapshot", "params": {"device": "/dev/video0"}}))
        resp = json.loads(await ws.recv())
        print("EXPIRED_TOKEN_RESPONSE:", resp)
        assert resp["error"]["code"] == -32001
        assert "expired" in resp["error"]["message"]

    await server.stop()


