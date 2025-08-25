# MediaMTX Camera Service - Go Implementation

**Version:** 1.0.0  
**Status:** Development  
**Technology Stack:** Go 1.19+  
**Migration Status:** From Python Implementation  

## Project Overview

This is the Go implementation of the MediaMTX Camera Service, providing high-performance camera discovery, streaming, recording, and snapshot capabilities. This implementation targets 5x performance improvement over the Python version while maintaining full API compatibility.

## Migration Protection Notice

**IMPORTANT:** This Go implementation is developed in parallel with the existing Python implementation. The original Python service in `../mediamtx-camera-service/` remains completely untouched and operational. This separation ensures:

- Zero risk to existing production deployments
- Clean development environment for Go implementation
- Ability to compare performance and functionality
- Gradual migration path with rollback capability

## Technology Stack

### Core Technologies
- **Language:** Go 1.19+
- **WebSocket:** gorilla/websocket
- **HTTP Server:** net/http (built-in)
- **Authentication:** golang.org/x/crypto/bcrypt, golang-jwt/jwt/v4
- **Configuration:** viper
- **Logging:** logrus
- **Testing:** testify

### Performance Targets
- **Response Time:** <100ms for 95% of requests (5x improvement)
- **Concurrency:** 1000+ simultaneous WebSocket connections (10x improvement)
- **Throughput:** 1000+ requests/second (5x improvement)
- **Resource Usage:** 50% reduction in memory and CPU

### Architecture Components
- **WebSocket JSON-RPC Server:** Real-time client communication
- **Camera Discovery Monitor:** USB camera detection and monitoring
- **MediaMTX Path Manager:** Dynamic stream management
- **Health & Monitoring:** System health and resource tracking
- **Security Middleware:** Authentication and authorization

## API Compatibility

This Go implementation maintains 100% API compatibility with the Python version:

- **WebSocket JSON-RPC 2.0:** Identical protocol and message formats
- **Authentication:** Same JWT and API key mechanisms
- **Methods:** All JSON-RPC methods implemented with identical signatures
- **Notifications:** Real-time event notifications with same payload structure
- **Error Codes:** Identical error codes and response formats

## Development Status

- [x] Project structure and documentation migration
- [ ] Core WebSocket server implementation
- [ ] Camera discovery and monitoring
- [ ] MediaMTX integration
- [ ] Authentication and security
- [ ] Performance optimization
- [ ] Testing and validation

## Quick Start

### Prerequisites
- Go 1.19 or higher
- MediaMTX server running
- USB cameras connected

### Build and Run
```bash
# Build the application
make build

# Run the server
make run

# Run tests
make test
```

### Configuration
The service uses the same configuration structure as the Python implementation, with Go-specific optimizations. See `docs/deployment/go-deployment-guide.md` for detailed setup instructions.

## Documentation

Complete documentation is available in the `docs/` directory:

- **Architecture:** `docs/architecture/overview.md` - System design and component architecture
- **API Reference:** `docs/api/json_rpc_methods.md` - Complete JSON-RPC API documentation
- **Development:** `docs/development/` - Coding standards, setup, and development guidelines
- **Deployment:** `docs/deployment/go-deployment-guide.md` - Production deployment guide
- **Requirements:** `docs/requirements/` - Functional and non-functional requirements

## Performance Comparison

| Metric | Python Implementation | Go Target | Improvement |
|--------|---------------------|-----------|-------------|
| API Response Time | <500ms | <100ms | 5x |
| Concurrent Connections | 50-100 | 1000+ | 10x |
| Requests/Second | 100-200 | 1000+ | 5x |
| Memory Usage | <80% | <60% | 25% reduction |
| CPU Usage | <70% | <50% | 30% reduction |

## Contributing

This project follows the established development principles and ground rules from the original Python implementation. See `docs/development/project-ground-rules.md` and `docs/development/roles-responsibilities.md` for development guidelines.

## License

Same license as the Python implementation. See LICENSE file for details.

---

**Note:** This Go implementation is under active development. For production use, refer to the stable Python implementation in `../mediamtx-camera-service/`.
