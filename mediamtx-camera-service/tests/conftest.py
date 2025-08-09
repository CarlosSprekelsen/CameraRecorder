import os

def pytest_sessionstart(session):
    # Provide a deterministic secret for JWT tests
    os.environ.setdefault("CAMERA_SERVICE_JWT_SECRET", "test-secret-key")
    os.environ.setdefault("CAMERA_SERVICE_RATE_RPM", "1000")

