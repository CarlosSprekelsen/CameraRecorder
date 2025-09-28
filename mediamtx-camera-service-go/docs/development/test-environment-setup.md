# Test Environment Setup Guide

**Version:** 1.0.0  
**Date:** 2025-09-28  
**Purpose:** Complete test environment setup for MediaMTX Camera Service development and testing

## **🎯 OVERVIEW**

This guide provides step-by-step instructions for setting up a complete test environment for the MediaMTX Camera Service, including all dependencies, configuration, and validation steps.

## **📋 PREREQUISITES**

### **System Requirements**
- **OS:** Ubuntu 20.04+ or equivalent Linux distribution
- **RAM:** Minimum 4GB (8GB recommended)
- **Storage:** 10GB free space
- **CPU:** 2+ cores
- **Network:** Internet connection for package downloads

### **Required Software**
- **Go:** 1.21+ (for building the service)
- **Docker:** 20.10+ (for MediaMTX container)
- **FFmpeg:** 4.4+ (for video processing)
- **V4L2:** Video4Linux2 support (for camera access)
- **Git:** 2.30+ (for source control)

## **🔧 INSTALLATION STEPS**

### **Step 1: Install System Dependencies**

```bash
# Update package lists
sudo apt update && sudo apt upgrade -y

# Install essential packages
sudo apt install -y \
    build-essential \
    git \
    curl \
    wget \
    vim \
    htop \
    netstat-nat \
    lsof \
    ffmpeg \
    v4l-utils \
    linux-headers-$(uname -r)

# Install Go (if not already installed)
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Install Docker (if not already installed)
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER
```

### **Step 2: Setup MediaMTX Service**

```bash
# Clone the repository
git clone https://github.com/your-org/mediamtx-camera-service-go.git
cd mediamtx-camera-service-go

# Install Go dependencies
go mod download
go mod tidy

# Build the service
go build -o camera-service ./cmd/camera-service
```

### **Step 3: Configure MediaMTX**

```bash
# Create MediaMTX configuration
sudo mkdir -p /opt/mediamtx/config
sudo cp config/mediamtx.yml /opt/mediamtx/config/

# Create service directories
sudo mkdir -p /opt/camera-service/{recordings,snapshots,logs}
sudo chown -R $USER:$USER /opt/camera-service
```

### **Step 4: Setup Test Cameras**

```bash
# Check available video devices
ls -la /dev/video*

# Test camera access
v4l2-ctl --list-devices

# Create test camera devices (if no real cameras)
sudo modprobe v4l2loopback devices=2
```

## **🧪 TEST ENVIRONMENT VALIDATION**

### **Validation Script**

Create `/tmp/validate-test-environment.sh`:

```bash
#!/bin/bash

echo "🔍 Validating Test Environment Setup..."

# Check Go installation
if ! command -v go &> /dev/null; then
    echo "❌ Go not installed"
    exit 1
fi
echo "✅ Go installed: $(go version)"

# Check Docker installation
if ! command -v docker &> /dev/null; then
    echo "❌ Docker not installed"
    exit 1
fi
echo "✅ Docker installed: $(docker --version)"

# Check FFmpeg installation
if ! command -v ffmpeg &> /dev/null; then
    echo "❌ FFmpeg not installed"
    exit 1
fi
echo "✅ FFmpeg installed: $(ffmpeg -version | head -1)"

# Check V4L2 support
if ! ls /dev/video* &> /dev/null; then
    echo "❌ No video devices found"
    exit 1
fi
echo "✅ Video devices found: $(ls /dev/video* | wc -l) devices"

# Check service build
if [ ! -f "./camera-service" ]; then
    echo "❌ Service not built"
    exit 1
fi
echo "✅ Service built successfully"

# Check configuration files
if [ ! -f "/opt/mediamtx/config/mediamtx.yml" ]; then
    echo "❌ MediaMTX config missing"
    exit 1
fi
echo "✅ MediaMTX configuration present"

# Check service directories
if [ ! -d "/opt/camera-service" ]; then
    echo "❌ Service directories missing"
    exit 1
fi
echo "✅ Service directories created"

echo "🎉 Test environment validation complete!"
```

### **Run Validation**

```bash
chmod +x /tmp/validate-test-environment.sh
/tmp/validate-test-environment.sh
```

## **🚀 QUICK START TESTING**

### **Start MediaMTX**

```bash
# Start MediaMTX container
docker run -d \
    --name mediamtx \
    -p 9997:9997 \
    -p 8554:8554 \
    -p 8888:8888 \
    -p 8889:8889 \
    -v /opt/mediamtx/config:/mediamtx/config \
    bluenviron/mediamtx:latest
```

### **Start Camera Service**

```bash
# Start the camera service
./camera-service -config config/ultra-efficient-edge.yaml

# In another terminal, test the service
curl http://localhost:8002/health
```

### **Run Integration Tests**

```bash
# Run all tests
go test -v ./...

# Run specific test suites
go test -v ./internal/websocket/
go test -v ./internal/mediamtx/
```

## **🔧 TROUBLESHOOTING**

### **Common Issues**

#### **Issue: No Video Devices**
```bash
# Solution: Create virtual cameras
sudo modprobe v4l2loopback devices=2
```

#### **Issue: Permission Denied**
```bash
# Solution: Add user to video group
sudo usermod -a -G video $USER
# Logout and login again
```

#### **Issue: MediaMTX Connection Failed**
```bash
# Solution: Check MediaMTX status
docker logs mediamtx
curl http://localhost:9997/v3/config/global/get
```

#### **Issue: Service Won't Start**
```bash
# Solution: Check configuration
./camera-service -config config/ultra-efficient-edge.yaml -validate
```

### **Debug Commands**

```bash
# Check service logs
journalctl -u camera-service -f

# Check MediaMTX logs
docker logs mediamtx -f

# Check port bindings
sudo netstat -tlnp | grep -E "(8002|9997|8554)"

# Check video devices
v4l2-ctl --list-devices
```

## **📊 PERFORMANCE TESTING**

### **Load Testing Setup**

```bash
# Install load testing tools
go install github.com/rakyll/hey@latest

# Run load tests
hey -n 1000 -c 10 http://localhost:8002/health
```

### **WebSocket Testing**

```bash
# Install WebSocket testing tools
npm install -g wscat

# Test WebSocket connection
wscat -c ws://localhost:8002/ws
```

## **🔒 SECURITY CONSIDERATIONS**

### **Test Environment Security**

- **Isolation:** Run tests in isolated environment
- **Credentials:** Use test-only credentials
- **Network:** Restrict network access to test environment
- **Data:** Use test data only, no production data

### **Security Testing**

```bash
# Test authentication
curl -X POST http://localhost:8002/ws \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"authenticate","params":{"auth_token":"test-token"}}'

# Test authorization
# (Add specific authorization tests)
```

## **📈 MONITORING AND METRICS**

### **Health Monitoring**

```bash
# Check service health
curl http://localhost:8003/health

# Check detailed health
curl http://localhost:8003/health/detailed

# Check readiness
curl http://localhost:8003/health/ready
```

### **Performance Metrics**

```bash
# Check system resources
htop

# Check network connections
ss -tuln | grep -E "(8002|8003|9997)"

# Check disk usage
df -h /opt/camera-service
```

## **🎯 NEXT STEPS**

1. **Complete Setup:** Follow all installation steps
2. **Run Validation:** Execute validation script
3. **Start Services:** Launch MediaMTX and camera service
4. **Run Tests:** Execute integration test suite
5. **Verify Functionality:** Test all API endpoints

## **📚 ADDITIONAL RESOURCES**

- **API Documentation:** `docs/api/json_rpc_methods.md`
- **Configuration Guide:** `docs/configuration/`
- **Troubleshooting:** `docs/troubleshooting/`
- **Performance Tuning:** `docs/performance/`

---

**This test environment setup ensures a complete, validated development and testing environment for the MediaMTX Camera Service.**
