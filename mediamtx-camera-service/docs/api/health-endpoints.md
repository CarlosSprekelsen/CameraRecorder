# Health Endpoints API Reference

**Version:** 1.0  
**Base URL:** `http://localhost:8003`  
**Content-Type:** `application/json`

## Overview

The MediaMTX Camera Service provides REST health endpoints for monitoring system status, component health, and Kubernetes readiness probes as specified in Architecture Decision AD-6.

## Endpoints

### GET `/health/system`

Returns overall system health status with detailed component information.

**Request:**
```http
GET /health/system
```

**Response:**
```json
{
  "status": "healthy|degraded|unhealthy",
  "timestamp": "2025-08-06T10:30:00.123456Z",
  "components": {
    "mediamtx": {
      "status": "healthy|unhealthy",
      "details": "MediaMTX controller is healthy"
    },
    "camera_monitor": {
      "status": "healthy|unhealthy", 
      "details": "Camera monitor is running with 2 cameras"
    },
    "service_manager": {
      "status": "healthy|unhealthy",
      "details": "Service manager is running"
    }
  }
}
```

**Status Codes:**
- `200 OK` - System is healthy or degraded
- `500 Internal Server Error` - System is unhealthy or error occurred

**Response Fields:**
- `status`: Overall system status (`healthy`, `degraded`, `unhealthy`)
- `timestamp`: ISO 8601 timestamp of health check
- `components`: Object containing health status of individual components

### GET `/health/cameras`

Returns camera discovery and monitoring system health.

**Request:**
```http
GET /health/cameras
```

**Response:**
```json
{
  "status": "healthy|unhealthy",
  "timestamp": "2025-08-06T10:30:00.123456Z",
  "details": "Camera monitor is running with 2 cameras"
}
```

**Status Codes:**
- `200 OK` - Camera system is healthy
- `500 Internal Server Error` - Camera system is unhealthy

**Response Fields:**
- `status`: Camera system status (`healthy`, `unhealthy`)
- `timestamp`: ISO 8601 timestamp of health check
- `details`: Human-readable description of camera system status

### GET `/health/mediamtx`

Returns MediaMTX server integration health status.

**Request:**
```http
GET /health/mediamtx
```

**Response:**
```json
{
  "status": "healthy|unhealthy",
  "timestamp": "2025-08-06T10:30:00.123456Z",
  "details": "MediaMTX controller is healthy"
}
```

**Status Codes:**
- `200 OK` - MediaMTX integration is healthy
- `500 Internal Server Error` - MediaMTX integration is unhealthy

**Response Fields:**
- `status`: MediaMTX integration status (`healthy`, `unhealthy`)
- `timestamp`: ISO 8601 timestamp of health check
- `details`: Human-readable description of MediaMTX status

### GET `/health/ready`

Kubernetes readiness probe endpoint.

**Request:**
```http
GET /health/ready
```

**Response (Ready):**
```json
{
  "status": "ready",
  "timestamp": "2025-08-06T10:30:00.123456Z"
}
```

**Response (Not Ready):**
```json
{
  "status": "not_ready",
  "timestamp": "2025-08-06T10:30:00.123456Z",
  "details": {
    "mediamtx": "unhealthy",
    "service_manager": "healthy"
  }
}
```

**Status Codes:**
- `200 OK` - Service is ready to receive traffic
- `503 Service Unavailable` - Service is not ready

**Readiness Criteria:**
- MediaMTX controller is healthy
- Service manager is healthy
- All critical components are operational

## Health Status Definitions

### System Status
- **healthy**: All components are operational
- **degraded**: Some components have issues but service is functional
- **unhealthy**: Critical components are failing

### Component Status
- **healthy**: Component is operational
- **unhealthy**: Component has issues or is not responding

## Error Responses

All endpoints may return error responses in the following format:

```json
{
  "status": "unhealthy",
  "timestamp": "2025-08-06T10:30:00.123456Z",
  "error": "Description of the error"
}
```

## Configuration

Health endpoints are configured via the security configuration:

```yaml
security:
  health:
    port: 8003
    bind_address: "0.0.0.0"
```

Environment variables:
- `HEALTH_PORT`: Health server port (default: 8003)
- `HEALTH_BIND_ADDRESS`: Health server bind address (default: 0.0.0.0)

## Monitoring Integration

### Prometheus Metrics

Health endpoints can be integrated with Prometheus monitoring:

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'camera-service-health'
    static_configs:
      - targets: ['localhost:8003']
    metrics_path: '/health/system'
```

### Kubernetes Health Checks

Example Kubernetes deployment with health checks:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: camera-service
spec:
  template:
    spec:
      containers:
      - name: camera-service
        image: camera-service:latest
        ports:
        - containerPort: 8002  # WebSocket
        - containerPort: 8003  # Health
        livenessProbe:
          httpGet:
            path: /health/system
            port: 8003
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8003
          initialDelaySeconds: 5
          periodSeconds: 5
```

## Usage Examples

### Check System Health
```bash
curl -X GET http://localhost:8003/health/system
```

### Check Camera System
```bash
curl -X GET http://localhost:8003/health/cameras
```

### Check MediaMTX Integration
```bash
curl -X GET http://localhost:8003/health/mediamtx
```

### Kubernetes Readiness Check
```bash
curl -X GET http://localhost:8003/health/ready
```

### Health Check with jq
```bash
curl -s http://localhost:8003/health/system | jq '.status'
```

## Troubleshooting

### Common Issues

1. **Health server not starting**
   - Check if port 8003 is available
   - Verify security configuration is valid
   - Check service manager logs

2. **Components showing unhealthy**
   - Verify MediaMTX server is running
   - Check camera monitor is operational
   - Review service manager status

3. **Kubernetes readiness probe failing**
   - Ensure all critical components are healthy
   - Check service dependencies are available
   - Verify network connectivity

### Debug Information

Enable debug logging to get detailed health check information:

```bash
# Set log level to DEBUG
export LOG_LEVEL=DEBUG

# Check health with verbose output
curl -v http://localhost:8003/health/system
```

## Security Considerations

- Health endpoints are designed for monitoring and should not expose sensitive information
- Consider network policies to restrict access to health endpoints
- Health data should not contain authentication tokens or secrets
- Use HTTPS in production environments

## Version History

- **v1.0**: Initial implementation with system, camera, MediaMTX, and readiness endpoints 