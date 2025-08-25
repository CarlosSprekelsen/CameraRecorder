# Go Development Setup

**Version:** 1.0  
**Authors:** Project Team  
**Date:** 2025-01-15  
**Status:** Approved  
**Related Epic/Story:** Go Implementation Setup  

**Purpose:**  
Provide comprehensive setup instructions for the Go development environment, including tooling, build process, and development workflow for the MediaMTX Camera Service Go implementation.

---

## 1. Prerequisites

### System Requirements
- **Operating System:** Linux (Ubuntu 20.04+), macOS, or Windows
- **Go Version:** 1.19 or higher
- **Memory:** Minimum 4GB RAM, 8GB recommended
- **Storage:** 2GB free space for Go toolchain and dependencies
- **Network:** Internet connection for dependency downloads

### Required Software
- **Go:** [Download from golang.org](https://golang.org/dl/)
- **Git:** Version control system
- **Make:** Build automation (usually pre-installed on Linux/macOS)
- **MediaMTX:** Media server for integration testing

---

## 2. Go Installation

### Linux (Ubuntu/Debian)
```bash
# Download and install Go
wget https://go.dev/dl/go1.19.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.19.linux-amd64.tar.gz

# Add Go to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
```

### macOS
```bash
# Using Homebrew
brew install go

# Or download from golang.org
# Download go1.19.darwin-amd64.pkg and install

# Verify installation
go version
```

### Windows
1. Download `go1.19.windows-amd64.msi` from [golang.org/dl](https://golang.org/dl/)
2. Run the installer and follow the prompts
3. Open Command Prompt and verify: `go version`

---

## 3. Development Environment Setup

### IDE Configuration
Recommended IDEs with Go support:

#### Visual Studio Code
1. Install VS Code from [code.visualstudio.com](https://code.visualstudio.com/)
2. Install Go extension: `golang.go`
3. Install additional extensions:
   - `ms-vscode.go` (Go language support)
   - `zxh404.vscode-proto3` (Protocol buffer support)
   - `ms-vscode.vscode-json` (JSON support)

#### GoLand (JetBrains)
1. Download from [jetbrains.com/go](https://www.jetbrains.com/go/)
2. Install and configure Go SDK
3. Enable automatic imports and formatting

### Go Tools Installation
```bash
# Install essential Go tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/go-delve/delve/cmd/dlv@latest
go install golang.org/x/tools/cmd/godoc@latest
go install github.com/ramya-rao-a/go-outline@latest
go install github.com/uudashr/gopkgs/v2/cmd/gopkgs@latest
go install github.com/nsf/gocode@latest
go install github.com/rogpeppe/godef@latest
go install github.com/sqs/goreturns@latest
go install github.com/golang/lint/golint@latest
```

### Linter Configuration
Create `.golangci.yml` in the project root:
```yaml
run:
  timeout: 5m
  go: "1.19"

linters:
  enable:
    - gofmt
    - goimports
    - govet
    - errcheck
    - staticcheck
    - gosimple
    - ineffassign
    - unused
    - misspell
    - gosec

linters-settings:
  gofmt:
    simplify: true
  goimports:
    local-prefixes: github.com/camerarecorder/mediamtx-camera-service-go

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - errcheck
```

---

## 4. Project Setup

### Clone Repository
```bash
# Clone the repository
git clone <repository-url>
cd mediamtx-camera-service-go

# Initialize Go module (if not already done)
go mod init github.com/camerarecorder/mediamtx-camera-service-go
```

### Install Dependencies
```bash
# Download and tidy dependencies
go mod download
go mod tidy

# Verify dependencies
go mod verify
```

### Build Configuration
```bash
# Build the application
make build

# Run the application
make run

# Run tests
make test
```

---

## 5. Development Workflow

### Code Formatting
```bash
# Format code automatically
make fmt

# Or use gofmt directly
gofmt -w .

# Use goimports for import organization
goimports -w .
```

### Linting and Validation
```bash
# Run linter
make lint

# Or use golangci-lint directly
golangci-lint run

# Run go vet
make vet
```

### Testing
```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test
go test ./internal/camera -v

# Run benchmarks
go test -bench=. ./internal/camera
```

### Debugging
```bash
# Debug with Delve
dlv debug cmd/server/main.go

# Or use VS Code debug configuration
# Create .vscode/launch.json:
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Package",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/cmd/server/main.go"
        }
    ]
}
```

---

## 6. Configuration Management

### Environment Setup
```bash
# Create environment file
cp config/config.example.yaml config/config.yaml

# Set environment variables
export CAMERA_SERVICE_PORT=8003
export CAMERA_SERVICE_WEBSOCKET_PORT=8002
export CAMERA_SERVICE_LOG_LEVEL=debug
```

### Configuration File
Create `config/config.yaml`:
```yaml
server:
  port: 8003
  websocket_port: 8002
  log_level: "info"

camera:
  discovery_interval: "5s"
  max_cameras: 16
  polling_enabled: true

mediamtx:
  api_url: "http://localhost:9997"
  rtsp_port: 8554
  webrtc_port: 8889
  hls_port: 8888

security:
  jwt_secret: "your-secret-key-here"
  token_expiry: "24h"
  max_login_attempts: 5
```

---

## 7. Integration Testing Setup

### MediaMTX Installation
```bash
# Download MediaMTX
wget https://github.com/bluenviron/mediamtx/releases/latest/download/mediamtx_linux_amd64.tar.gz
tar -xzf mediamtx_linux_amd64.tar.gz
sudo mv mediamtx /usr/local/bin/

# Create MediaMTX configuration
cat > mediamtx.yml << EOF
paths:
  all:
    runOnDemand: ffmpeg -f v4l2 -i /dev/video0 -c:v libx264 -preset ultrafast -b:v 600k -f rtsp rtsp://127.0.0.1:8554/camera0

api: yes
apiAddress: :9997
rtspAddress: :8554
webrtcAddress: :8889
hlsAddress: :8888
EOF

# Start MediaMTX
mediamtx mediamtx.yml
```

### Test Camera Setup
```bash
# Install v4l2 utilities for camera testing
sudo apt-get install v4l-utils

# List available cameras
v4l2-ctl --list-devices

# Test camera capabilities
v4l2-ctl -d /dev/video0 --list-formats-ext
```

---

## 8. Performance Profiling

### CPU Profiling
```bash
# Run with CPU profiling
go run -cpuprofile=cpu.prof cmd/server/main.go

# Analyze profile
go tool pprof cpu.prof
```

### Memory Profiling
```bash
# Run with memory profiling
go run -memprofile=mem.prof cmd/server/main.go

# Analyze profile
go tool pprof mem.prof
```

### HTTP Profiling
```bash
# Add profiling endpoints to main.go
import _ "net/http/pprof"

# Access profiles at runtime
curl http://localhost:8003/debug/pprof/heap
curl http://localhost:8003/debug/pprof/goroutine
```

---

## 9. Continuous Integration

### GitHub Actions Setup
Create `.github/workflows/go.yml`:
```yaml
name: Go

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.19
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run tests
      run: go test -v ./...
    
    - name: Run linter
      run: golangci-lint run
    
    - name: Build
      run: go build -v ./cmd/server
```

---

## 10. Troubleshooting

### Common Issues

#### Go Module Issues
```bash
# Clear module cache
go clean -modcache

# Reset go.mod
rm go.mod go.sum
go mod init github.com/camerarecorder/mediamtx-camera-service-go
go mod tidy
```

#### Permission Issues
```bash
# Fix camera device permissions
sudo usermod -a -G video $USER
sudo chmod 666 /dev/video*

# Restart session or reboot
```

#### Port Conflicts
```bash
# Check port usage
sudo netstat -tlnp | grep :8002
sudo netstat -tlnp | grep :8003

# Kill conflicting processes
sudo kill -9 <PID>
```

#### Build Issues
```bash
# Clean build artifacts
make clean

# Rebuild with verbose output
go build -v -x ./cmd/server
```

---

## 11. Development Best Practices

### Code Organization
- Keep related functionality in the same package
- Use interfaces for dependency injection
- Implement proper error handling
- Write comprehensive tests

### Git Workflow
```bash
# Create feature branch
git checkout -b feature/websocket-server

# Make changes and commit
git add .
git commit -m "feat: implement WebSocket JSON-RPC server [Story:E1/S2]"

# Push and create pull request
git push origin feature/websocket-server
```

### Documentation
- Update README.md for new features
- Document API changes in `docs/api/`
- Add examples for new functionality
- Keep TODO comments tracked in roadmap

---

## 12. Performance Monitoring

### Metrics Collection
```bash
# Run with metrics enabled
go run -tags=metrics cmd/server/main.go

# Monitor with Prometheus
# Add prometheus.yml configuration
```

### Logging Configuration
```bash
# Set log level
export LOG_LEVEL=debug

# Enable structured logging
export LOG_FORMAT=json
```

---

**Next Steps:**
1. Complete the setup process
2. Run initial build and tests
3. Configure IDE for optimal development experience
4. Set up integration testing environment
5. Begin implementation following the coding standards

**References:**
- [Go Documentation](https://golang.org/doc/)
- [Go Modules](https://golang.org/ref/mod)
- [Go Testing](https://golang.org/pkg/testing/)
- [Go Profiling](https://golang.org/pkg/runtime/pprof/)
