# MediaMTX Camera Service

A lightweight WebSocket JSON-RPC 2.0 service that provides real-time USB camera monitoring and control using MediaMTX as the media server backend.

## 🎯 Project Goals

### Primary Objectives
- **Real-time Camera Discovery**: Sub-200ms USB camera connect/disconnect detection
- **WebSocket JSON-RPC 2.0 API**: Standards-compliant camera control interface  
- **MediaMTX Integration**: Leverage proven media server for streaming/recording
- **Production-Ready**: Native Linux deployment with systemd services
- **Lightweight**: Minimal wrapper (~30MB RAM) around robust MediaMTX core

### Problem & Solution
Create a **thin wrapper** around a proven recomended solution like MediaMTX that adds:
- Real-time USB camera monitoring
- WebSocket JSON-RPC 2.0 API
- Dynamic camera stream management
- Production deployment automation

## 📁 Project Structure

```
mediamtx-camera-service/
├── src/                           # Python source code
│   ├── camera_service/            # Main application
│   ├── mediamtx_wrapper/          # MediaMTX REST API client
│   ├── websocket_server/          # WebSocket JSON-RPC server
│   ├── camera_discovery/          # USB camera monitoring
│   └── common/                    # Shared utilities
├── config/                        # Configuration files
├── deployment/                    # Systemd services & scripts
├── docs/                          # Documentation
├── tests/                         # Test suite
├── examples/                      # Usage examples
└── tools/                         # Development tools
```

## 📚 Documentation

- **[Architecture Overview](docs/architecture/overview.md)** - System design and component interaction
- **[API Reference](docs/api/json-rpc-methods.md)** - Complete JSON-RPC method specification
- **[Installation Guide](docs/deployment/installation.md)** - Production deployment instructions
- **[Development Setup](docs/development/setup.md)** - Development environment configuration

## ✅ TODO List - Professional Development Approach

### Phase 1: Analysis & Architecture Validation (Week 1)

#### 📋 Requirements & Architecture Review
- [ ] **Study Generated Project Structure**
  - [ ] Review all generated files and documentation
  - [ ] Validate project organization and naming conventions
  - [ ] Identify missing components or structural issues
  - [ ] Document any required changes to generated structure

- [ ] **Architecture Analysis**
  - [X] Review MediaMTX capabilities and limitations ([docs/architecture/overview.md](docs/architecture/overview.md))
  - [ ] Validate MediaMTX REST API compatibility and versioning
  - [ ] Analyze camera discovery approach (polling vs event-driven)
  - [ ] Evaluate WebSocket vs HTTP API trade-offs
  - [ ] Assess performance requirements and bottlenecks

- [ ] **Technical Feasibility Study**
  - [ ] Test MediaMTX installation and basic operation
  - [ ] Validate USB camera detection with v4l2-ctl
  - [ ] Prototype MediaMTX REST API communication
  - [ ] Verify WebSocket JSON-RPC 2.0 library compatibility
  - [ ] Test systemd service deployment approach

#### 🔍 Design Decisions & Documentation
- [ ] **API Design Validation**
  - [ ] Review JSON-RPC method specifications ([docs/api/json-rpc-methods.md](docs/api/json-rpc-methods.md))
  - [ ] Validate error handling approach and error codes
  - [ ] Define data models and serialization formats
  - [ ] Specify notification schemas and timing

- [ ] **Configuration Management**
  - [ ] Review configuration file structure and validation
  - [ ] Define environment variable override strategy  
  - [ ] Plan configuration hot-reload capabilities
  - [ ] Document configuration migration strategy

- [ ] **Risk Assessment**
  - [ ] Identify potential architecture bottlenecks
  - [ ] Analyze MediaMTX dependency risks and mitigation
  - [ ] Evaluate USB device permission and security implications
  - [ ] Plan error recovery and graceful degradation strategies

### Phase 2: Proof of Concept (Week 2)
- [ ] **MediaMTX Integration PoC**
  - [ ] Basic REST API client implementation
  - [ ] Camera stream creation/deletion workflow
  - [ ] Recording start/stop integration

- [ ] **Camera Discovery PoC**
  - [ ] USB device monitoring implementation
  - [ ] v4l2-ctl capability detection integration
  - [ ] Event notification system

- [ ] **WebSocket JSON-RPC PoC**
  - [ ] Basic server with core methods (`ping`, `get_camera_list`)
  - [ ] Real-time notification broadcasting
  - [ ] Error handling and client management

### Phase 3: Core Implementation (Week 3-4)
- [ ] **Complete API Implementation**
  - [ ] All JSON-RPC methods per specification
  - [ ] Comprehensive error handling
  - [ ] Performance optimization

- [ ] **Testing & Validation**
  - [ ] Unit test coverage (>80%)
  - [ ] Integration testing with real hardware
  - [ ] Performance benchmarking and optimization
  - [ ] Load testing and stress testing

### Phase 4: Production Readiness (Week 5-6)
- [ ] **Deployment & Operations**
  - [ ] Production installation automation
  - [ ] Monitoring and health checks
  - [ ] Documentation completion
  - [ ] Security hardening and validation

## 🎯 Decision Points Requiring Review

### Critical Architecture Decisions
1. **MediaMTX Version Compatibility** - Which MediaMTX version to target?
2. **Camera Discovery Method** - Polling vs udev events vs hybrid approach?
3. **Configuration Management** - YAML vs JSON vs environment variables priority?
4. **Error Recovery Strategy** - How to handle MediaMTX failures and restarts?
5. **API Versioning** - How to handle future API changes and backward compatibility?

### Implementation Choices Needing Validation
1. **WebSocket vs HTTP** - Confirm WebSocket-only approach vs hybrid REST+WebSocket
2. **Authentication Strategy** - None, JWT, API keys, or client certificates?
3. **Logging Approach** - Structured JSON logs vs traditional format?
4. **Performance Targets** - Define acceptable latency and throughput requirements
5. **Resource Limits** - Memory, CPU, and storage constraints

### Deployment Strategy Review
1. **System Integration** - Systemd vs other service managers?
2. **User Permissions** - Dedicated user vs existing users for camera access?
3. **File System Layout** - Validate `/opt/camera-service/` approach
4. **Update Strategy** - In-place updates vs blue-green deployment?
5. **Backup & Recovery** - Configuration and recording backup strategy

## 🔧 Development Commands

```bash
make dev-install    # Set up development environment
make test          # Run test suite
make lint          # Code quality checks
make format        # Auto-format code
make run           # Start development server
make clean         # Clean build artifacts
```

## 📊 Success Metrics

### Week 1 Goals
- [ ] MediaMTX REST API client working
- [ ] USB camera detection functional
- [ ] Basic WebSocket server responding to `ping`
- [ ] Real-time camera connect/disconnect notifications

### Week 2 Goals  
- [ ] Complete JSON-RPC API implemented
- [ ] Recording and snapshot functionality
- [ ] Comprehensive test coverage (>80%)
- [ ] Documentation complete

### Week 3 Goals
- [ ] Production deployment working
- [ ] Systemd services stable
- [ ] Performance benchmarks met (<200ms camera detection)
- [ ] Client examples functional

## 🤝 Contributing

1. Review [Contributing Guidelines](CONTRIBUTING.md)
2. Check [Development Setup](docs/development/setup.md)  
3. Follow [Coding Standards](docs/development/coding-standards.md)
4. Run tests: `make test`
5. Submit pull request

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

---

**Status**: 🚧 Active Development  
**Version**: 0.1.0  
**Python**: 3.10+  
**Target**: Ubuntu 22.04+ Linux
