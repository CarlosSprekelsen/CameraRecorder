# MediaMTX Camera Service

A lightweight WebSocket JSON-RPC 2.0 service that provides real-time USB camera monitoring and control using MediaMTX as the media server backend.


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


## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

---

**Status**: 🚧 Active Development  
**Version**: 0.1.0  
**Python**: 3.10+  
**Target**: Ubuntu 22.04+ Linux
