# Test Naming Conventions

## Standard Test Function Naming Format

All test functions must follow this standardized naming pattern:

```go
func Test{Component}_{Method}_{Requirement}_{Scenario}(t *testing.T)
func Benchmark{Component}_{Method}_{Scenario}(b *testing.B)
```

### Components:
- `SnapshotManager` - Snapshot management functionality
- `StreamManager` - Stream management functionality  
- `RecordingManager` - Recording management functionality
- `Controller` - MediaMTX controller functionality
- `Client` - MediaMTX client functionality
- `WebSocketServer` - WebSocket server functionality
- `WebSocketMethods` - WebSocket method implementations
- `CameraMonitor` - Camera monitoring functionality
- `PathValidator` - Path validation functionality
- `ConfigManager` - Configuration management
- `JWTHandler` - JWT authentication
- `RoleManager` - Role-based access control

### Methods/Actions:
- `New` - Constructor/creation tests
- `Start` - Startup functionality
- `Stop` - Shutdown functionality
- `Get` - Retrieval operations
- `Set` - Setting operations
- `Create` - Creation operations
- `Delete` - Deletion operations
- `Update` - Update operations
- `Validate` - Validation operations
- `Process` - Processing operations

### Requirements:
- `ReqMTX001` - MediaMTX service integration
- `ReqMTX002` - Stream management capabilities
- `ReqMTX003` - Path creation and deletion
- `ReqMTX004` - Health monitoring
- `ReqMTX007` - Error handling and recovery
- `ReqAPI001` - WebSocket JSON-RPC 2.0 API endpoint
- `ReqAPI002` - JSON-RPC 2.0 protocol implementation
- `ReqAPI003` - Request/response message handling
- `ReqARCH001` - Progressive Readiness Pattern compliance
- `ReqSEC001` - Authentication and authorization
- `ReqCAM001` - Camera hardware integration

### Scenarios (Optional):
- `Success` - Happy path testing
- `ErrorHandling` - Error condition testing
- `Concurrent` - Concurrency testing
- `Integration` - Integration testing
- `Performance` - Performance testing
- `Tier0` - Direct hardware access
- `Tier1` - USB direct capture
- `Tier2` - RTSP immediate capture
- `Tier3` - RTSP stream activation
- `MultiTier` - Multi-tier integration
- `RealHardware` - Real hardware testing

## Examples:

### Good Examples:
```go
func TestSnapshotManager_New_ReqMTX001_Success(t *testing.T)
func TestSnapshotManager_TakeSnapshot_ReqMTX002_Success(t *testing.T)
func TestSnapshotManager_TakeSnapshot_ReqMTX002_ErrorHandling(t *testing.T)
func TestSnapshotManager_TakeSnapshot_ReqMTX002_Tier1_USBDirect(t *testing.T)
func TestWebSocketServer_Start_ReqAPI001_Success(t *testing.T)
func TestController_GetHealth_ReqMTX004_Success(t *testing.T)
func BenchmarkSnapshotManager_TakeSnapshot_Performance(b *testing.B)
```

### Subtests for Complex Scenarios:
```go
func TestSnapshotManager_MultiTier_ReqMTX002(t *testing.T) {
    t.Run("Tier1_USBDirect", func(t *testing.T) { ... })
    t.Run("Tier2_RTSPImmediate", func(t *testing.T) { ... })
    t.Run("Tier3_RTSPActivation", func(t *testing.T) { ... })
}
```

## Migration Status:

### ✅ **Completed Files:**
- `internal/mediamtx/snapshot_manager_test.go` - All 15 test functions standardized
- `internal/mediamtx/client_test.go` - All 13 test functions standardized  
- `internal/mediamtx/controller_test.go` - 10 test functions standardized (28 total)
- `internal/websocket/server_test.go` - 5 test functions standardized (24 total)
- `internal/websocket/methods_test.go` - 5 test functions standardized (23 total)

### 🔄 **In Progress:**
- Remaining controller test functions (18 remaining)
- Remaining websocket test functions (42 remaining)
- Camera test files (8 files)
- Security test files (6 files)
- Config test files (3 files)
- Integration test files (3 files)
- Performance test files (3 files)

### 📊 **Progress Summary:**
- **Total Test Files:** 65
- **Files Partially/Fully Updated:** 5
- **Test Functions Standardized:** ~38 out of ~400+
- **Completion:** ~10%

## Migration Rules:

1. **All new tests** must follow the standard format
2. **Existing tests** should be renamed during maintenance  
3. **Requirement tags** are mandatory for all tests
4. **Scenario suffixes** should be used when multiple test cases exist for the same method
5. **Subtests** should be used for complex multi-scenario testing

## Before/After Examples:

### Before (Inconsistent):
```go
func TestNewSnapshotManager_ReqMTX001(t *testing.T)
func TestWebSocketServer_Creation(t *testing.T) 
func TestController_ConcurrentAccess_ReqMTX001(t *testing.T)
```

### After (Standardized):
```go
func TestSnapshotManager_New_ReqMTX001_Success(t *testing.T)
func TestWebSocketServer_New_ReqAPI001_Success(t *testing.T)
func TestController_GetHealth_ReqMTX001_Concurrent(t *testing.T)
```

## Enforcement:

This naming convention should be enforced through:
- Code review guidelines
- Automated linting rules  
- CI/CD pipeline checks
- Pre-commit hooks
