# Go Deployment Guide

**Version:** 1.0  
**Authors:** Project Team  
**Date:** 2025-01-15  
**Status:** Approved  
**Related Epic/Story:** Go Implementation Deployment  

**Purpose:**  
Provide comprehensive deployment instructions for the MediaMTX Camera Service Go implementation, including SystemD service configuration, binary deployment, and container-based deployment options.

---

## 1. Production Deployment Overview

### Deployment Options
- **Binary Deployment:** Direct binary installation with SystemD service
- **Container Deployment:** Docker container with orchestration
- **Cloud Deployment:** Kubernetes deployment for scalable environments

### System Requirements
- **Operating System:** Linux (Ubuntu 20.04+ recommended)
- **Go Runtime:** Not required (statically linked binary)
- **Memory:** 512MB minimum, 2GB recommended
- **Storage:** 1GB for binary and configuration, additional for recordings
- **Network:** Ports 8002 (WebSocket), 8003 (HTTP) available

---

## 2. Binary Deployment

### Build Production Binary
```bash
# Build optimized production binary
make build-prod

# Or build manually
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
  -ldflags="-w -s" \
  -o build/mediamtx-camera-service-go \
  cmd/server/main.go

# Verify binary
file build/mediamtx-camera-service-go
# Should show: ELF 64-bit LSB executable, statically linked
```

### Installation Script
Create `deployment/scripts/install.sh`:
```bash
#!/bin/bash

set -e

# Configuration
SERVICE_NAME="mediamtx-camera-service-go"
INSTALL_DIR="/opt/camera-service"
CONFIG_DIR="/etc/camera-service"
SERVICE_USER="camera-service"
BINARY_NAME="mediamtx-camera-service-go"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Installing MediaMTX Camera Service (Go Implementation)...${NC}"

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo -e "${RED}This script must be run as root${NC}"
   exit 1
fi

# Create service user
if ! id "$SERVICE_USER" &>/dev/null; then
    echo "Creating service user: $SERVICE_USER"
    useradd -r -s /bin/false -d "$INSTALL_DIR" "$SERVICE_USER"
fi

# Create directories
echo "Creating installation directories..."
mkdir -p "$INSTALL_DIR"
mkdir -p "$CONFIG_DIR"
mkdir -p "$INSTALL_DIR/logs"
mkdir -p "$INSTALL_DIR/recordings"
mkdir -p "$INSTALL_DIR/snapshots"

# Copy binary
echo "Installing binary..."
cp "build/$BINARY_NAME" "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/$BINARY_NAME"

# Copy configuration
echo "Installing configuration..."
cp config/config.yaml "$CONFIG_DIR/"
chown -R "$SERVICE_USER:$SERVICE_USER" "$CONFIG_DIR"

# Set ownership
chown -R "$SERVICE_USER:$SERVICE_USER" "$INSTALL_DIR"

# Create SystemD service
echo "Creating SystemD service..."
cat > /etc/systemd/system/$SERVICE_NAME.service << EOF
[Unit]
Description=MediaMTX Camera Service (Go Implementation)
After=network.target
Wants=network.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/$BINARY_NAME
ExecReload=/bin/kill -HUP \$MAINPID
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=$SERVICE_NAME

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$INSTALL_DIR $CONFIG_DIR

# Resource limits
LimitNOFILE=65536
LimitNPROC=4096

# Environment variables
Environment=CONFIG_PATH=$CONFIG_DIR/config.yaml
Environment=LOG_LEVEL=info

[Install]
WantedBy=multi-user.target
EOF

# Reload SystemD and enable service
echo "Enabling and starting service..."
systemctl daemon-reload
systemctl enable $SERVICE_NAME
systemctl start $SERVICE_NAME

# Verify installation
if systemctl is-active --quiet $SERVICE_NAME; then
    echo -e "${GREEN}Installation completed successfully!${NC}"
    echo -e "${YELLOW}Service status:${NC}"
    systemctl status $SERVICE_NAME --no-pager
else
    echo -e "${RED}Installation failed!${NC}"
    echo -e "${YELLOW}Service logs:${NC}"
    journalctl -u $SERVICE_NAME --no-pager -n 20
    exit 1
fi

echo -e "${GREEN}MediaMTX Camera Service is now running on:${NC}"
echo -e "  WebSocket: ws://localhost:8002/ws"
echo -e "  HTTP Health: http://localhost:8003/health"
echo -e "  Logs: journalctl -u $SERVICE_NAME -f"
```

### Uninstallation Script
Create `deployment/scripts/uninstall.sh`:
```bash
#!/bin/bash

set -e

# Configuration
SERVICE_NAME="mediamtx-camera-service-go"
INSTALL_DIR="/opt/camera-service"
CONFIG_DIR="/etc/camera-service"
SERVICE_USER="camera-service"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}Uninstalling MediaMTX Camera Service (Go Implementation)...${NC}"

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo -e "${RED}This script must be run as root${NC}"
   exit 1
fi

# Stop and disable service
echo "Stopping service..."
systemctl stop $SERVICE_NAME || true
systemctl disable $SERVICE_NAME || true

# Remove SystemD service file
echo "Removing SystemD service..."
rm -f /etc/systemd/system/$SERVICE_NAME.service
systemctl daemon-reload

# Remove installation files
echo "Removing installation files..."
rm -rf "$INSTALL_DIR"
rm -rf "$CONFIG_DIR"

# Remove service user (optional - uncomment if desired)
# echo "Removing service user..."
# userdel -r "$SERVICE_USER" || true

echo -e "${GREEN}Uninstallation completed successfully!${NC}"
```

---

## 3. SystemD Service Management

### Service Configuration
The SystemD service file includes:
- **User isolation:** Runs as dedicated service user
- **Security hardening:** No new privileges, protected system
- **Resource limits:** File descriptors and process limits
- **Automatic restart:** Restarts on failure with backoff
- **Logging:** Integrated with system journal

### Service Commands
```bash
# Start service
sudo systemctl start mediamtx-camera-service-go

# Stop service
sudo systemctl stop mediamtx-camera-service-go

# Restart service
sudo systemctl restart mediamtx-camera-service-go

# Check status
sudo systemctl status mediamtx-camera-service-go

# View logs
sudo journalctl -u mediamtx-camera-service-go -f

# Enable auto-start
sudo systemctl enable mediamtx-camera-service-go

# Disable auto-start
sudo systemctl disable mediamtx-camera-service-go
```

### Service Monitoring
```bash
# Monitor service health
sudo systemctl is-active mediamtx-camera-service-go

# Check resource usage
sudo systemctl show mediamtx-camera-service-go --property=MemoryCurrent,CPUUsageNSec

# View recent logs
sudo journalctl -u mediamtx-camera-service-go --since "1 hour ago"
```

---

## 4. Configuration Management

### Production Configuration
Create `/etc/camera-service/config.yaml`:
```yaml
server:
  port: 8003
  websocket_port: 8002
  log_level: "info"
  log_format: "json"

camera:
  discovery_interval: "5s"
  max_cameras: 16
  polling_enabled: true
  udev_enabled: true

mediamtx:
  api_url: "http://localhost:9997"
  rtsp_port: 8554
  webrtc_port: 8889
  hls_port: 8888
  timeout: "10s"
  retry_attempts: 3

security:
  jwt_secret: "your-production-secret-key-here"
  token_expiry: "24h"
  max_login_attempts: 5
  rate_limit_requests: 100
  rate_limit_window: "1m"

storage:
  recordings_path: "/opt/camera-service/recordings"
  snapshots_path: "/opt/camera-service/snapshots"
  max_file_size: "2GB"
  retention_days: 30
  cleanup_interval: "1h"

monitoring:
  metrics_enabled: true
  health_check_interval: "30s"
  prometheus_port: 9090
```

### Environment Variables
```bash
# Set in SystemD service or environment file
export CAMERA_SERVICE_CONFIG_PATH=/etc/camera-service/config.yaml
export CAMERA_SERVICE_LOG_LEVEL=info
export CAMERA_SERVICE_LOG_FORMAT=json
export CAMERA_SERVICE_JWT_SECRET=your-secret-key
```

---

## 5. Container Deployment

### Dockerfile
Create `Dockerfile`:
```dockerfile
# Build stage
FROM golang:1.19-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o mediamtx-camera-service-go \
    cmd/server/main.go

# Production stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create service user
RUN addgroup -g 1001 -S camera-service && \
    adduser -u 1001 -S camera-service -G camera-service

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/mediamtx-camera-service-go .

# Create directories
RUN mkdir -p /app/config /app/logs /app/recordings /app/snapshots && \
    chown -R camera-service:camera-service /app

# Switch to service user
USER camera-service

# Expose ports
EXPOSE 8002 8003

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8003/health || exit 1

# Run binary
CMD ["./mediamtx-camera-service-go"]
```

### Docker Compose
Create `docker-compose.yml`:
```yaml
version: '3.8'

services:
  mediamtx-camera-service:
    build: .
    container_name: mediamtx-camera-service-go
    restart: unless-stopped
    ports:
      - "8002:8002"  # WebSocket
      - "8003:8003"  # HTTP
    volumes:
      - ./config:/app/config:ro
      - ./recordings:/app/recordings
      - ./snapshots:/app/snapshots
      - ./logs:/app/logs
    environment:
      - CONFIG_PATH=/app/config/config.yaml
      - LOG_LEVEL=info
    depends_on:
      - mediamtx
    networks:
      - camera-network

  mediamtx:
    image: aler9/mediamtx:latest
    container_name: mediamtx
    restart: unless-stopped
    ports:
      - "8554:8554"  # RTSP
      - "8889:8889"  # WebRTC
      - "8888:8888"  # HLS
      - "9997:9997"  # API
    volumes:
      - ./mediamtx.yml:/mediamtx.yml:ro
    command: mediamtx /mediamtx.yml
    networks:
      - camera-network

networks:
  camera-network:
    driver: bridge

volumes:
  recordings:
  snapshots:
  logs:
```

### Kubernetes Deployment
Create `k8s/deployment.yaml`:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mediamtx-camera-service
  labels:
    app: mediamtx-camera-service
spec:
  replicas: 1
  selector:
    matchLabels:
      app: mediamtx-camera-service
  template:
    metadata:
      labels:
        app: mediamtx-camera-service
    spec:
      containers:
      - name: mediamtx-camera-service
        image: mediamtx-camera-service-go:latest
        ports:
        - containerPort: 8002
          name: websocket
        - containerPort: 8003
          name: http
        env:
        - name: CONFIG_PATH
          value: "/app/config/config.yaml"
        - name: LOG_LEVEL
          value: "info"
        volumeMounts:
        - name: config
          mountPath: /app/config
          readOnly: true
        - name: recordings
          mountPath: /app/recordings
        - name: snapshots
          mountPath: /app/snapshots
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8003
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8003
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: mediamtx-camera-service-config
      - name: recordings
        persistentVolumeClaim:
          claimName: recordings-pvc
      - name: snapshots
        persistentVolumeClaim:
          claimName: snapshots-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: mediamtx-camera-service
spec:
  selector:
    app: mediamtx-camera-service
  ports:
  - name: websocket
    port: 8002
    targetPort: 8002
  - name: http
    port: 8003
    targetPort: 8003
  type: ClusterIP
```

---

## 6. Security Configuration

### Firewall Configuration
```bash
# UFW (Ubuntu)
sudo ufw allow 8002/tcp  # WebSocket
sudo ufw allow 8003/tcp  # HTTP Health
sudo ufw allow 8554/tcp  # RTSP
sudo ufw allow 8889/tcp  # WebRTC
sudo ufw allow 8888/tcp  # HLS

# iptables
sudo iptables -A INPUT -p tcp --dport 8002 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 8003 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 8554 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 8889 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 8888 -j ACCEPT
```

### SSL/TLS Configuration
```bash
# Generate SSL certificate
sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /etc/ssl/private/camera-service.key \
  -out /etc/ssl/certs/camera-service.crt

# Configure nginx reverse proxy
sudo nano /etc/nginx/sites-available/camera-service
```

Nginx configuration:
```nginx
server {
    listen 443 ssl;
    server_name your-domain.com;

    ssl_certificate /etc/ssl/certs/camera-service.crt;
    ssl_certificate_key /etc/ssl/private/camera-service.key;

    location /ws {
        proxy_pass http://localhost:8002;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location / {
        proxy_pass http://localhost:8003;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

---

## 7. Monitoring and Logging

### Log Management
```bash
# Configure log rotation
sudo nano /etc/logrotate.d/mediamtx-camera-service

# Log rotation configuration
/var/log/camera-service/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 camera-service camera-service
    postrotate
        systemctl reload mediamtx-camera-service-go
    endscript
}
```

### Prometheus Metrics
```yaml
# Add to config.yaml
monitoring:
  metrics_enabled: true
  prometheus_port: 9090
  metrics_path: "/metrics"
```

### Health Check Endpoints
```bash
# Check service health
curl http://localhost:8003/health

# Check detailed status
curl http://localhost:8003/health/detailed

# Check camera status
curl http://localhost:8003/health/cameras

# Check MediaMTX status
curl http://localhost:8003/health/mediamtx
```

---

## 8. Backup and Recovery

### Backup Script
Create `deployment/scripts/backup.sh`:
```bash
#!/bin/bash

# Configuration
BACKUP_DIR="/backup/camera-service"
SERVICE_NAME="mediamtx-camera-service-go"
CONFIG_DIR="/etc/camera-service"
INSTALL_DIR="/opt/camera-service"

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Create timestamp
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Backup configuration
echo "Backing up configuration..."
tar -czf "$BACKUP_DIR/config_$TIMESTAMP.tar.gz" -C /etc camera-service

# Backup recordings and snapshots
echo "Backing up media files..."
tar -czf "$BACKUP_DIR/media_$TIMESTAMP.tar.gz" \
  -C /opt camera-service/recordings camera-service/snapshots

# Backup service configuration
echo "Backing up service configuration..."
cp /etc/systemd/system/$SERVICE_NAME.service "$BACKUP_DIR/service_$TIMESTAMP.service"

# Create backup manifest
cat > "$BACKUP_DIR/manifest_$TIMESTAMP.txt" << EOF
Backup created: $(date)
Service: $SERVICE_NAME
Configuration: config_$TIMESTAMP.tar.gz
Media files: media_$TIMESTAMP.tar.gz
Service config: service_$TIMESTAMP.service
EOF

echo "Backup completed: $BACKUP_DIR"
```

### Recovery Script
Create `deployment/scripts/recover.sh`:
```bash
#!/bin/bash

# Configuration
BACKUP_DIR="/backup/camera-service"
SERVICE_NAME="mediamtx-camera-service-go"

echo "Recovering MediaMTX Camera Service..."

# Stop service
systemctl stop $SERVICE_NAME

# Restore configuration
echo "Restoring configuration..."
tar -xzf "$BACKUP_DIR/config_*.tar.gz" -C /etc

# Restore media files
echo "Restoring media files..."
tar -xzf "$BACKUP_DIR/media_*.tar.gz" -C /opt

# Restore service configuration
echo "Restoring service configuration..."
cp "$BACKUP_DIR/service_*.service" /etc/systemd/system/$SERVICE_NAME.service

# Reload and start service
systemctl daemon-reload
systemctl start $SERVICE_NAME

echo "Recovery completed!"
```

---

## 9. Performance Tuning

### System Tuning
```bash
# Increase file descriptor limits
echo "* soft nofile 65536" >> /etc/security/limits.conf
echo "* hard nofile 65536" >> /etc/security/limits.conf

# Optimize network settings
echo "net.core.somaxconn = 65535" >> /etc/sysctl.conf
echo "net.ipv4.tcp_max_syn_backlog = 65535" >> /etc/sysctl.conf

# Apply changes
sysctl -p
```

### Application Tuning
```yaml
# Performance configuration in config.yaml
performance:
  max_goroutines: 1000
  worker_pool_size: 16
  connection_timeout: "30s"
  read_timeout: "15s"
  write_timeout: "15s"
  idle_timeout: "60s"
```

---

## 10. Troubleshooting

### Common Issues

#### Service Won't Start
```bash
# Check service status
systemctl status mediamtx-camera-service-go

# View detailed logs
journalctl -u mediamtx-camera-service-go -n 50

# Check configuration
/opt/camera-service/mediamtx-camera-service-go --config-test
```

#### Permission Issues
```bash
# Fix file permissions
chown -R camera-service:camera-service /opt/camera-service
chown -R camera-service:camera-service /etc/camera-service

# Fix camera device permissions
sudo usermod -a -G video camera-service
sudo chmod 666 /dev/video*
```

#### Port Conflicts
```bash
# Check port usage
netstat -tlnp | grep :8002
netstat -tlnp | grep :8003

# Kill conflicting processes
sudo kill -9 $(lsof -t -i:8002)
sudo kill -9 $(lsof -t -i:8003)
```

#### Memory Issues
```bash
# Check memory usage
free -h
ps aux | grep mediamtx-camera-service-go

# Monitor with htop
htop -p $(pgrep mediamtx-camera-service-go)
```

---

**Next Steps:**
1. Choose deployment method (binary, container, or cloud)
2. Follow installation instructions for selected method
3. Configure security and monitoring
4. Test deployment and verify functionality
5. Set up backup and recovery procedures

**References:**
- [SystemD Documentation](https://systemd.io/)
- [Docker Documentation](https://docs.docker.com/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Nginx Documentation](https://nginx.org/en/docs/)
