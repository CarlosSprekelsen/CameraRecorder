# SSL/TLS Setup Guide

**Version:** 1.0  
**Architecture Decision:** AD-7  
**Implementation:** Sprint 1 (S6)

## Overview

The MediaMTX Camera Service supports SSL/TLS encryption for secure WebSocket connections (WSS) as specified in Architecture Decision AD-7. This guide covers certificate generation, configuration, and deployment.

## SSL Configuration

### Security Configuration Schema

```yaml
security:
  ssl:
    enabled: true
    cert_file: "/path/to/cert.pem"
    key_file: "/path/to/key.pem"
    verify_mode: "CERT_REQUIRED"
```

### Environment Variables

- `SSL_ENABLED`: Enable SSL/TLS (true/false)
- `SSL_CERT_FILE`: Path to SSL certificate file
- `SSL_KEY_FILE`: Path to SSL private key file

## Certificate Types

### Self-Signed Certificates (Development)

For development and testing environments, self-signed certificates provide encryption without requiring a Certificate Authority (CA).

### CA-Signed Certificates (Production)

For production environments, use certificates signed by a trusted Certificate Authority for proper browser and client validation.

## Certificate Generation

### Self-Signed Certificate Generation

#### Using OpenSSL

```bash
# Generate private key
openssl genrsa -out camera-service.key 2048

# Generate certificate signing request (CSR)
openssl req -new -key camera-service.key -out camera-service.csr

# Generate self-signed certificate
openssl x509 -req -days 365 -in camera-service.csr -signkey camera-service.key -out camera-service.crt

# Combine certificate and key for WebSocket server
cat camera-service.crt camera-service.key > camera-service.pem
```

#### Using Let's Encrypt (Production)

```bash
# Install certbot
sudo apt install certbot

# Generate certificate for your domain
sudo certbot certonly --standalone -d your-domain.com

# Copy certificates to service location
sudo cp /etc/letsencrypt/live/your-domain.com/fullchain.pem /opt/camera-service/ssl/
sudo cp /etc/letsencrypt/live/your-domain.com/privkey.pem /opt/camera-service/ssl/
```

### Certificate Management Script

Create a certificate management script for automated renewal:

```bash
#!/bin/bash
# /opt/camera-service/scripts/ssl-setup.sh

CERT_DIR="/opt/camera-service/ssl"
SERVICE_USER="camera-service"

# Create SSL directory
sudo mkdir -p $CERT_DIR
sudo chown $SERVICE_USER:$SERVICE_USER $CERT_DIR
sudo chmod 700 $CERT_DIR

# Generate self-signed certificate
openssl req -x509 -newkey rsa:2048 -keyout $CERT_DIR/camera-service.key \
    -out $CERT_DIR/camera-service.crt -days 365 -nodes \
    -subj "/C=US/ST=State/L=City/O=Organization/CN=localhost"

# Combine certificate and key
cat $CERT_DIR/camera-service.crt $CERT_DIR/camera-service.key > $CERT_DIR/camera-service.pem

# Set proper permissions
sudo chown $SERVICE_USER:$SERVICE_USER $CERT_DIR/*
sudo chmod 600 $CERT_DIR/*
```

## Configuration Examples

### Development Configuration

```yaml
# /opt/camera-service/config/camera-service.yaml
security:
  ssl:
    enabled: true
    cert_file: "/opt/camera-service/ssl/camera-service.pem"
    key_file: "/opt/camera-service/ssl/camera-service.key"
```

### Production Configuration

```yaml
# /opt/camera-service/config/camera-service.yaml
security:
  ssl:
    enabled: true
    cert_file: "/opt/camera-service/ssl/fullchain.pem"
    key_file: "/opt/camera-service/ssl/privkey.pem"
```

### Environment Variable Configuration

```bash
# Set environment variables
export SSL_ENABLED=true
export SSL_CERT_FILE="/opt/camera-service/ssl/camera-service.pem"
export SSL_KEY_FILE="/opt/camera-service/ssl/camera-service.key"
```

## WebSocket SSL Implementation

### SSL Context Creation

The WebSocket server creates an SSL context for secure connections:

```python
import ssl

def create_ssl_context(cert_file: str, key_file: str) -> ssl.SSLContext:
    """Create SSL context for WebSocket server."""
    ssl_context = ssl.create_default_context(ssl.Purpose.CLIENT_AUTH)
    ssl_context.load_cert_chain(cert_file, key_file)
    ssl_context.verify_mode = ssl.CERT_REQUIRED
    return ssl_context
```

### WebSocket Server SSL Integration

```python
import websockets

async def start_ssl_websocket_server(host: str, port: int, ssl_context: ssl.SSLContext):
    """Start WebSocket server with SSL."""
    server = await websockets.serve(
        handle_connection,
        host,
        port,
        ssl=ssl_context
    )
    return server
```

## Client Connection Examples

### Python Client (WSS)

```python
import ssl
import websockets
import json

# SSL context for client
ssl_context = ssl.create_default_context()
ssl_context.check_hostname = False
ssl_context.verify_mode = ssl.CERT_NONE  # For self-signed certificates

# Connect to secure WebSocket
async with websockets.connect(
    "wss://localhost:8002/ws",
    ssl=ssl_context
) as websocket:
    # Send authenticated request
    request = {
        "jsonrpc": "2.0",
        "method": "get_camera_list",
        "id": 1,
        "params": {"auth_token": "your-jwt-token"}
    }
    
    await websocket.send(json.dumps(request))
    response = await websocket.recv()
    print(json.loads(response))
```

### JavaScript Client (WSS)

```javascript
// Connect to secure WebSocket
const ws = new WebSocket("wss://localhost:8002/ws");

ws.onopen = function() {
    // Send authenticated request
    const request = {
        jsonrpc: "2.0",
        method: "get_camera_list",
        id: 1,
        params: { auth_token: "your-jwt-token" }
    };
    
    ws.send(JSON.stringify(request));
};

ws.onmessage = function(event) {
    const response = JSON.parse(event.data);
    console.log(response);
};
```

### cURL Testing

```bash
# Test SSL connection with cURL
curl -k -i -N -H "Connection: Upgrade" \
    -H "Upgrade: websocket" \
    -H "Sec-WebSocket-Version: 13" \
    -H "Sec-WebSocket-Key: x3JJHMbDL1EzLkh9GBhXDw==" \
    https://localhost:8002/ws
```

## Certificate Validation

### Certificate Verification

```python
import ssl
import socket

def verify_certificate(hostname: str, port: int, cert_file: str):
    """Verify SSL certificate."""
    context = ssl.create_default_context()
    context.load_verify_locations(cert_file)
    
    with socket.create_connection((hostname, port)) as sock:
        with context.wrap_socket(sock, server_hostname=hostname) as ssock:
            cert = ssock.getpeercert()
            print(f"Certificate subject: {cert['subject']}")
            print(f"Certificate issuer: {cert['issuer']}")
            return cert
```

### Certificate Information

```bash
# View certificate information
openssl x509 -in camera-service.crt -text -noout

# Check certificate expiry
openssl x509 -in camera-service.crt -noout -dates

# Verify certificate chain
openssl verify camera-service.crt
```

## Security Best Practices

### Certificate Security

1. **Strong Private Keys**
   ```bash
   # Generate 4096-bit RSA key for production
   openssl genrsa -out camera-service.key 4096
   ```

2. **Secure File Permissions**
   ```bash
   # Set restrictive permissions
   chmod 600 /opt/camera-service/ssl/*
   chown camera-service:camera-service /opt/camera-service/ssl/*
   ```

3. **Certificate Rotation**
   - Implement automated certificate renewal
   - Monitor certificate expiry dates
   - Plan for certificate replacement

### Network Security

1. **Firewall Configuration**
   ```bash
   # Allow HTTPS/WSS traffic
   sudo ufw allow 8002/tcp  # WebSocket SSL
   sudo ufw allow 8003/tcp  # Health endpoints
   ```

2. **Reverse Proxy Setup**
   ```nginx
   # Nginx configuration
   server {
       listen 443 ssl;
       server_name your-domain.com;
       
       ssl_certificate /path/to/cert.pem;
       ssl_certificate_key /path/to/key.pem;
       
       location /ws {
           proxy_pass http://localhost:8002;
           proxy_http_version 1.1;
           proxy_set_header Upgrade $http_upgrade;
           proxy_set_header Connection "upgrade";
           proxy_set_header Host $host;
       }
   }
   ```

## Troubleshooting

### Common SSL Issues

1. **Certificate Not Found**
   ```bash
   # Check certificate file exists
   ls -la /opt/camera-service/ssl/
   
   # Verify certificate is readable
   openssl x509 -in /opt/camera-service/ssl/camera-service.pem -text -noout
   ```

2. **Permission Denied**
   ```bash
   # Fix file permissions
   sudo chown camera-service:camera-service /opt/camera-service/ssl/*
   sudo chmod 600 /opt/camera-service/ssl/*
   ```

3. **Certificate Expired**
   ```bash
   # Check certificate expiry
   openssl x509 -in /opt/camera-service/ssl/camera-service.pem -noout -dates
   
   # Regenerate certificate
   ./ssl-setup.sh
   ```

### SSL Connection Testing

```bash
# Test SSL connection
openssl s_client -connect localhost:8002 -servername localhost

# Test certificate chain
openssl verify -CAfile /path/to/ca-bundle.crt /path/to/cert.pem
```

### Debug SSL Issues

Enable SSL debugging:

```bash
# Set SSL debug environment variable
export SSLKEYLOGFILE=/tmp/ssl.log

# Check SSL logs
tail -f /opt/camera-service/logs/camera-service.log | grep -i ssl
```

## Monitoring and Maintenance

### Certificate Monitoring

1. **Expiry Monitoring**
   ```bash
   # Check certificate expiry
   openssl x509 -in /opt/camera-service/ssl/camera-service.pem -noout -enddate
   
   # Set up monitoring alert
   echo "Certificate expires in $(openssl x509 -in /opt/camera-service/ssl/camera-service.pem -noout -enddate)"
   ```

2. **Automated Renewal**
   ```bash
   # Cron job for Let's Encrypt renewal
   0 12 * * * /usr/bin/certbot renew --quiet && systemctl reload camera-service
   ```

### SSL Metrics

Monitor SSL connection metrics:
- Successful SSL handshakes
- Failed SSL connections
- Certificate validation errors
- SSL protocol versions used

## Production Deployment

### Let's Encrypt Integration

```bash
# Install certbot
sudo apt install certbot

# Generate certificate
sudo certbot certonly --standalone -d your-domain.com

# Set up automatic renewal
sudo crontab -e
# Add: 0 12 * * * /usr/bin/certbot renew --quiet
```

### Certificate Authority Setup

For internal deployments, set up a private CA:

```bash
# Generate CA private key
openssl genrsa -out ca.key 4096

# Generate CA certificate
openssl req -new -x509 -days 3650 -key ca.key -out ca.crt

# Generate server certificate
openssl req -new -key server.key -out server.csr
openssl x509 -req -days 365 -in server.csr -CA ca.crt -CAkey ca.key -out server.crt
```

## Version History

- **v1.0**: Initial implementation with SSL/TLS support for WebSocket connections 