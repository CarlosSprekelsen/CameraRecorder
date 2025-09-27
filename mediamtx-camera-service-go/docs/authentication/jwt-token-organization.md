# 🔑 JWT Token Organization Guide

## 📁 **Proper File Structure**

This document explains the proper organization of JWT tokens and related files in the MediaMTX Camera Service project.

**⚠️ IMPORTANT**: This system uses JWT tokens for authentication, not API keys. All documentation and scripts have been updated to reflect this.

---

## 🛠️ **JWT Token Management Scripts**

### **Server-Side Management**

#### **Main Script: `generate-jwt-tokens.sh`**
**Location**: `mediamtx-camera-service-go/deployment/scripts/generate-jwt-tokens.sh`  
**Purpose**: Generate JWT tokens for all roles with configurable expiry

**Usage:**
```bash
cd mediamtx-camera-service-go/deployment/scripts
./generate-jwt-tokens.sh --expiry-hours 72
./generate-jwt-tokens.sh --role admin --expiry-hours 24
./generate-jwt-tokens.sh --help
```

**Options:**
- `--expiry-hours HOURS` - Token expiry in hours (default: 48)
- `--secret-key KEY` - JWT secret key (default: from config)
- `--help` - Show help message

**Generated Files:**
- JWT tokens for all roles (viewer, operator, admin)
- Environment variables for client testing
- JSON configuration files for fixtures

#### **Orchestration Script: `reinstall-with-tokens.sh`**
**Location**: `mediamtx-camera-service-go/deployment/scripts/reinstall-with-tokens.sh`  
**Purpose**: Complete system reinstall with fresh JWT token generation

**Usage:**
```bash
cd mediamtx-camera-service-go/deployment/scripts
sudo ./reinstall-with-tokens.sh
```

**Process:**
1. Uninstall existing service
2. Install fresh service
3. Generate fresh JWT tokens (72h expiry)
4. Setup test environment
5. Start service
6. Verify installation

#### **JWT Generator Tool: `cmd/jwt-generator/main.go`**
**Location**: `mediamtx-camera-service-go/cmd/jwt-generator/main.go`  
**Purpose**: Command-line JWT token generation utility

**Usage:**
```bash
cd mediamtx-camera-service-go
go run cmd/jwt-generator/main.go --role admin --expiry-hours 72 --format json
```

**Options:**
- `--role`: User role (viewer, operator, admin)
- `--expiry-hours`: Token expiry in hours (default: 48)
- `--secret-key`: JWT secret key (default: from config)
- `--user-id`: Custom user ID (default: test_<role>)
- `--format`: Output format (token, json)

---

## 📂 **File Locations**

### **Server Files**
```
mediamtx-camera-service-go/
├── deployment/scripts/
│   ├── generate-jwt-tokens.sh           # ✅ JWT token generation
│   ├── reinstall-with-tokens.sh         # ✅ Complete reinstall
│   ├── install.sh                       # ✅ Server installation
│   └── uninstall.sh                     # ✅ Server uninstallation
├── cmd/jwt-generator/
│   └── main.go                          # ✅ JWT generator tool
├── config/test/jwt-tokens/
│   └── jwt-tokens.json                  # ✅ Test JWT tokens
└── docs/authentication/
    ├── jwt-token-guide.md               # ✅ JWT documentation
    └── jwt-token-organization.md        # ✅ This file
```

### **Client Files**
```
MediaMTX-Camera-Service-Client/client/
├── tests/fixtures/
│   └── test_jwt_tokens.json             # ✅ Client test JWT tokens
├── scripts/
│   └── setup-test-keys.sh               # ✅ Client test setup
└── .test_env                            # ✅ Client test environment
```

---

## 🔄 **Workflow**

### **Development Workflow**
1. **Generate tokens**: `./generate-jwt-tokens.sh --expiry-hours 72`
2. **Test authentication**: Client loads tokens from `.test_env`
3. **Run tests**: `npm run test:integration`

### **Production Workflow**
1. **Complete reinstall**: `sudo ./reinstall-with-tokens.sh`
2. **Verify installation**: Check service status and health endpoints
3. **Test client**: Run integration tests with fresh tokens

### **Token Lifecycle**
1. **Generation**: Create JWT tokens with specified role and expiry
2. **Distribution**: Tokens copied to client environment files
3. **Authentication**: Client uses tokens for WebSocket authentication
4. **Expiration**: Tokens automatically expire at specified time
5. **Rotation**: Generate new tokens before expiration

---

## 🔧 **Configuration**

### **Server Configuration**
JWT settings in `/opt/camera-service/config/default.yaml`:
```yaml
security:
  jwt_secret_key: "edge-device-secret-key-change-in-production"
  jwt_expiry_hours: 48
```

### **Client Configuration**
Environment variables in `.test_env`:
```bash
export CAMERA_SERVICE_HOST=localhost
export CAMERA_SERVICE_PORT=8002
export CAMERA_SERVICE_WS_PATH=/ws
export TEST_ADMIN_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

---

## 🚨 **Security Notes**

### **Token Security**
- **Secret Key**: Use strong, unique secret keys in production
- **Expiry**: Set appropriate expiry times (24-72 hours)
- **Storage**: Never commit tokens to version control
- **Rotation**: Regenerate tokens regularly

### **File Permissions**
- **Scripts**: Executable by deployment user
- **Tokens**: Readable by client application
- **Config**: Protected from unauthorized access

### **Environment Separation**
- **Development**: Use test tokens with short expiry
- **Staging**: Use staging-specific secret keys
- **Production**: Use production-grade secret keys

---

## 📚 **Documentation**

### **Related Documents**
- `jwt-token-guide.md` - Detailed JWT token usage guide
- `../api/json_rpc_methods.md` - API authentication documentation
- `../security/api-key-management.md` - Security architecture (updated for JWT)

### **Migration Notes**
This system migrated from API keys to JWT tokens. Key differences:
- **Storage**: JWT tokens are self-contained (no database storage)
- **Validation**: Cryptographic signature validation
- **Revocation**: Wait for expiration (no immediate revocation)
- **Performance**: Faster validation (no database lookup)

---

## ✅ **Verification Checklist**

### **Token Generation**
- [ ] JWT generator tool works correctly
- [ ] Tokens contain all required claims (user_id, role, iat, exp)
- [ ] Tokens are signed with correct secret key
- [ ] Expiry times are set correctly

### **Client Integration**
- [ ] Environment variables are loaded correctly
- [ ] Client can authenticate with JWT tokens
- [ ] Authentication flow works end-to-end
- [ ] Integration tests pass with JWT tokens

### **Documentation**
- [ ] All documentation reflects JWT authentication
- [ ] API documentation is accurate
- [ ] Security documentation is updated
- [ ] Migration notes are clear

---

**Last Updated**: 2025-09-27  
**Version**: 1.0  
**Status**: Current Implementation
