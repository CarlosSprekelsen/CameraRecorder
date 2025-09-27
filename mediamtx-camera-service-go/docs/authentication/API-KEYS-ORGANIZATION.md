# 🔑 JWT Token Organization Guide

## 📁 **Proper File Structure**

This document explains the proper organization of JWT tokens and related files in the MediaMTX Camera Service project.

**⚠️ IMPORTANT**: This system now uses JWT tokens for authentication, not API keys. This document has been updated to reflect the current implementation.

### ✅ **Current Structure**

```
/home/carlossprekelsen/CameraRecorder/
├── .gitignore                                    # Excludes sensitive files
├── mediamtx-camera-service-go/                   # Server project
│   ├── deployment/
│   │   └── scripts/
│   │       ├── generate-jwt-tokens.sh           # ✅ JWT token generation
│   │       ├── reinstall-with-tokens.sh         # ✅ Complete reinstall with tokens
│   │       ├── install.sh                       # ✅ Server installation
│   │       └── uninstall.sh                     # ✅ Server uninstallation
│   ├── cmd/
│   │   └── jwt-generator/
│   │       └── main.go                          # ✅ JWT token generator tool
│   ├── config/
│   │   ├── default.yaml                         # ✅ Server configuration
│   │   └── test/                                # ✅ Test configurations
│   │       └── jwt-tokens/
│   │           └── jwt-tokens.json              # ✅ Test JWT tokens
│   ├── docs/
│   │   └── authentication/
│   │       ├── jwt-token-guide.md               # ✅ JWT token documentation
│   │       └── API-KEYS-ORGANIZATION.md         # ✅ This file (updated)
│   └── tests/
│       └── tools/
│           └── setup_test_environment.sh        # ✅ Test environment setup
│
├── MediaMTX-Camera-Service-Client/               # Client project
│   └── client/
│       ├── tests/
│       │   └── fixtures/
│       │       └── test_jwt_tokens.json         # ✅ Client test JWT tokens
│       ├── scripts/
│       │   └── setup-test-keys.sh               # ✅ Client test setup
│       └── .test_env                             # ✅ Client test environment (JWT tokens)
│
└── deployment/                                   # Deployment artifacts
    └── keys/                                     # Production tokens (gitignored)
        ├── production/                           # Production JWT tokens
        ├── staging/                              # Staging JWT tokens
        └── development/                          # Development JWT tokens
```

---

## 🛠️ **JWT Token Management Scripts**

### **Server-Side Management**

#### **Main Script: `generate-jwt-tokens.sh`**
Location: `mediamtx-camera-service-go/deployment/scripts/manage-api-keys.sh`

**Usage:**
```bash
cd mediamtx-camera-service-go
./deployment/scripts/manage-api-keys.sh <command> [environment]
```

**Commands:**
- `generate test` - Generate test API keys
- `generate production` - Generate production API keys
- `install test` - Install test keys to server
- `install production` - Install production keys to server
- `backup` - Backup existing API keys
- `test` - Test API key authentication
- `clean` - Clean temporary files

**Examples:**
```bash
# Generate and install test keys
./deployment/scripts/manage-api-keys.sh generate test
./deployment/scripts/manage-api-keys.sh install test

# Generate production keys (keep secure!)
./deployment/scripts/manage-api-keys.sh generate production
./deployment/scripts/manage-api-keys.sh install production
```

### **Client-Side Management**

#### **Client Setup Script: `setup-test-keys.sh`**
Location: `MediaMTX-Camera-Service-Client/client/scripts/setup-test-keys.sh`

**Usage:**
```bash
cd MediaMTX-Camera-Service-Client/client
./scripts/setup-test-keys.sh
```

**What it does:**
1. Checks for server test keys
2. Copies keys to client fixtures
3. Updates client `.test_env` file
4. Installs client dependencies
5. Runs basic integration tests

---

## 🔒 **Security Considerations**

### **File Permissions**
- **Server API keys**: `600` (read/write by owner only)
- **Test keys**: `644` (readable by group)
- **Production keys**: `600` (strictly confidential)

### **Git Exclusion**
The `.gitignore` file excludes:
- All API key files (`**/api-keys.json`, `**/*_keys.json`)
- Environment files (`**/*.env`, `.test_env`)
- Production deployment keys (`deployment/keys/production/`)

### **Key Formats**
- **Server Format**: `csk_` prefix with base64url encoding (32 bytes)
- **Environment**: Standard environment variable format
- **JSON**: Structured with metadata (created_at, expires_at, etc.)

---

## 🚀 **Quick Start Guide**

### **For Development/Testing:**

1. **Generate test keys:**
   ```bash
   cd mediamtx-camera-service-go
   ./deployment/scripts/manage-api-keys.sh generate test
   ```

2. **Install to server:**
   ```bash
   ./deployment/scripts/manage-api-keys.sh install test
   sudo systemctl restart camera-service
   ```

3. **Setup client:**
   ```bash
   cd ../MediaMTX-Camera-Service-Client/client
   ./scripts/setup-test-keys.sh
   ```

4. **Test authentication:**
   ```bash
   npm run test:integration -- --testPathPattern="authenticated_functionality"
   ```

### **For Production:**

1. **Generate production keys:**
   ```bash
   cd mediamtx-camera-service-go
   ./deployment/scripts/manage-api-keys.sh generate production
   ```

2. **Secure the keys:**
   ```bash
   # Keys are automatically stored in deployment/keys/production/
   # Ensure proper backup and access controls
   ```

3. **Install to server:**
   ```bash
   ./deployment/scripts/manage-api-keys.sh install production
   sudo systemctl restart camera-service
   ```

---

## 📋 **File Responsibilities**

### **Server Files:**
- `manage-api-keys.sh` - Main API key management
- `config/test/api-keys/` - Test keys for development
- `deployment/keys/` - Production keys (gitignored)

### **Client Files:**
- `setup-test-keys.sh` - Client test setup automation
- `tests/fixtures/test_api_keys.json` - Client test keys
- `.test_env` - Client test environment variables

### **Root Files:**
- `.gitignore` - Security exclusions
- `API-KEYS-ORGANIZATION.md` - This documentation

---

## ✅ **Benefits of This Organization**

1. **Security**: Sensitive files are properly excluded from version control
2. **Separation**: Clear distinction between test and production keys
3. **Automation**: Scripts handle key generation and deployment
4. **Consistency**: Standardized key formats and locations
5. **Maintainability**: Easy to find and manage API keys
6. **Documentation**: Clear guidance for developers and operations

---

## 🔧 **Troubleshooting**

### **Common Issues:**

1. **"API keys not found"**
   - Run: `./deployment/scripts/manage-api-keys.sh generate test`

2. **"Authentication failed"**
   - Install keys: `./deployment/scripts/manage-api-keys.sh install test`
   - Restart service: `sudo systemctl restart camera-service`

3. **"Client tests fail"**
   - Setup client: `./scripts/setup-test-keys.sh`
   - Check environment: `source .test_env`

4. **"Permission denied"**
   - Check file permissions: `ls -la /opt/camera-service/api-keys.json`
   - Fix ownership: `sudo chown camera-service:camera-service /opt/camera-service/api-keys.json`

---

## 📞 **Support**

For issues with API key management:
1. Check this documentation
2. Run the appropriate script with verbose output
3. Check server logs: `sudo journalctl -u camera-service -f`
4. Verify file permissions and ownership
