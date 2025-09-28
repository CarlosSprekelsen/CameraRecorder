# API Versioning Policy

**Version:** 1.0.0  
**Date:** 2025-09-28  
**Purpose:** Comprehensive API versioning strategy for MediaMTX Camera Service

## **🎯 OVERVIEW**

This document establishes the official API versioning policy for the MediaMTX Camera Service, ensuring backward compatibility, clear migration paths, and predictable evolution of the API.

## **📋 VERSIONING STRATEGY**

### **Semantic Versioning (SemVer)**

The API follows [Semantic Versioning 2.0.0](https://semver.org/) with the format: `MAJOR.MINOR.PATCH`

- **MAJOR:** Breaking changes that require client updates
- **MINOR:** New features that are backward compatible
- **PATCH:** Bug fixes that are backward compatible

### **Version Indicators**

#### **API Version Header**
```http
X-API-Version: 1.2.3
```

#### **JSON-RPC Response Metadata**
```json
{
  "jsonrpc": "2.0",
  "result": {...},
  "id": 1,
  "meta": {
    "api_version": "1.2.3",
    "deprecated": false,
    "sunset_date": null
  }
}
```

#### **WebSocket Connection Metadata**
```json
{
  "type": "connection_established",
  "api_version": "1.2.3",
  "supported_versions": ["1.0.0", "1.1.0", "1.2.3"],
  "deprecation_warnings": []
}
```

## **🔄 VERSION LIFECYCLE**

### **Version States**

| State | Description | Duration | Client Action Required |
|-------|-------------|----------|----------------------|
| **Current** | Latest stable version | Indefinite | None |
| **Deprecated** | Will be removed | 12 months | Plan migration |
| **Sunset** | Removed from service | N/A | Must upgrade |
| **Legacy** | Old versions still supported | 24 months | Consider upgrade |

### **Version Timeline**

```
Timeline: 0 ────────────── 12 months ────────────── 24 months
         │                    │                        │
    Current v1.0.0    Deprecated v1.0.0        Legacy v1.0.0
    New v1.1.0        Current v1.1.0        Deprecated v1.1.0
                         New v1.2.0              Current v1.2.0
```

## **📊 COMPATIBILITY MATRIX**

### **Backward Compatibility Rules**

| Change Type | MAJOR | MINOR | PATCH | Example |
|-------------|-------|-------|-------|---------|
| **Add Method** | ❌ | ✅ | ✅ | New `get_metrics` method |
| **Add Parameter** | ❌ | ✅ | ✅ | Optional `timeout` parameter |
| **Remove Method** | ✅ | ❌ | ❌ | Remove `old_method` |
| **Change Response** | ✅ | ❌ | ❌ | Change `result` structure |
| **Add Error Code** | ❌ | ✅ | ✅ | New `-32050` error |
| **Remove Error Code** | ✅ | ❌ | ❌ | Remove `-32010` error |
| **Change Authentication** | ✅ | ❌ | ❌ | JWT → OAuth2 |

### **Client Compatibility**

| Client Version | Server v1.0.0 | Server v1.1.0 | Server v1.2.0 | Server v2.0.0 |
|----------------|---------------|---------------|---------------|---------------|
| **v1.0.0** | ✅ Full | ✅ Full | ✅ Full | ❌ Breaking |
| **v1.1.0** | ✅ Full | ✅ Full | ✅ Full | ❌ Breaking |
| **v1.2.0** | ✅ Full | ✅ Full | ✅ Full | ❌ Breaking |
| **v2.0.0** | ⚠️ Limited | ⚠️ Limited | ⚠️ Limited | ✅ Full |

## **🚨 BREAKING CHANGES**

### **Definition of Breaking Changes**

A change is considered **breaking** if it:

1. **Removes** an existing method or parameter
2. **Changes** the structure of a response
3. **Modifies** the behavior of existing functionality
4. **Requires** client code changes to continue working

### **Breaking Change Examples**

#### **❌ Breaking: Remove Method**
```json
// v1.0.0 - Method exists
{"method": "get_old_metrics", "params": {}}

// v2.0.0 - Method removed
{"error": {"code": -32601, "message": "Method not found"}}
```

#### **❌ Breaking: Change Response Structure**
```json
// v1.0.0 - Old structure
{
  "result": {
    "cameras": [...],
    "count": 5
  }
}

// v2.0.0 - New structure (BREAKING)
{
  "result": {
    "data": {
      "cameras": [...],
      "total_count": 5
    }
  }
}
```

#### **✅ Non-Breaking: Add Optional Parameter**
```json
// v1.0.0 - Original method
{"method": "get_camera_list", "params": {}}

// v1.1.0 - Add optional parameter (NON-BREAKING)
{"method": "get_camera_list", "params": {"include_offline": true}}
```

## **📢 DEPRECATION PROCESS**

### **Deprecation Timeline**

#### **Phase 1: Announcement (12 months before removal)**
```json
{
  "jsonrpc": "2.0",
  "result": {...},
  "meta": {
    "deprecated": true,
    "deprecation_date": "2024-01-01",
    "sunset_date": "2025-01-01",
    "replacement": "new_method",
    "migration_guide": "https://docs.example.com/migration"
  }
}
```

#### **Phase 2: Warning Phase (6 months before removal)**
```json
{
  "jsonrpc": "2.0",
  "result": {...},
  "warnings": [
    {
      "type": "deprecation_warning",
      "message": "Method 'old_method' is deprecated and will be removed on 2025-01-01",
      "replacement": "new_method"
    }
  ]
}
```

#### **Phase 3: Removal (After 12 months)**
- Method returns `-32601 Method not found`
- Documentation updated
- Migration tools provided

### **Deprecation Communication**

1. **Release Notes:** Document in changelog
2. **API Documentation:** Mark as deprecated
3. **Client Libraries:** Update with warnings
4. **Migration Guides:** Provide step-by-step instructions
5. **Support:** Assist with migration

## **🔧 MIGRATION SUPPORT**

### **Migration Tools**

#### **Automated Migration Script**
```bash
# Install migration tool
npm install -g @camera-service/migration-tool

# Run migration
camera-service-migrate --from 1.0.0 --to 1.2.0 --config config.yaml
```

#### **Compatibility Checker**
```bash
# Check client compatibility
camera-service-compat --client-version 1.0.0 --server-version 1.2.0
```

### **Migration Guides**

#### **v1.0.0 → v1.1.0 Migration**
```markdown
## Breaking Changes: None
## New Features:
- Added `get_metrics` method
- Added optional `timeout` parameter to `take_snapshot`
## Action Required: None
```

#### **v1.2.0 → v2.0.0 Migration**
```markdown
## Breaking Changes:
- Removed `get_old_metrics` method
- Changed `get_camera_list` response structure
## Migration Steps:
1. Replace `get_old_metrics` with `get_metrics`
2. Update response parsing for `get_camera_list`
3. Test all functionality
```

## **📈 VERSION SUPPORT MATRIX**

### **Supported Versions**

| Version | Release Date | Support Until | Status |
|---------|--------------|---------------|---------|
| **v2.0.0** | 2025-01-01 | 2027-01-01 | Current |
| **v1.2.0** | 2024-06-01 | 2026-06-01 | Deprecated |
| **v1.1.0** | 2024-03-01 | 2026-03-01 | Legacy |
| **v1.0.0** | 2024-01-01 | 2026-01-01 | Legacy |

### **Version Support Policy**

- **Current:** Full support, all features
- **Deprecated:** Security fixes only, 12-month notice
- **Legacy:** Critical security fixes only, 24-month support
- **Unsupported:** No support, upgrade required

## **🧪 TESTING STRATEGY**

### **Compatibility Testing**

```bash
# Test client compatibility
go test -v ./tests/compatibility/

# Test version negotiation
go test -v ./tests/versioning/

# Test migration scenarios
go test -v ./tests/migration/
```

### **Version Negotiation Testing**

```go
func TestVersionNegotiation(t *testing.T) {
    // Test client requests specific version
    client := NewClient("1.1.0")
    response := client.Call("get_camera_list")
    
    // Verify server responds with compatible version
    assert.Equal(t, "1.1.0", response.Meta.APIVersion)
}
```

## **📋 IMPLEMENTATION CHECKLIST**

### **For Each Release**

- [ ] **Version Number:** Update according to SemVer
- [ ] **Changelog:** Document all changes
- [ ] **Breaking Changes:** Identify and document
- [ ] **Deprecation:** Mark deprecated features
- [ ] **Migration Guide:** Create if needed
- [ ] **Compatibility:** Test with existing clients
- [ ] **Documentation:** Update API docs
- [ ] **Client Libraries:** Update if needed

### **For Breaking Changes**

- [ ] **12-Month Notice:** Announce deprecation
- [ ] **Migration Tools:** Provide automated tools
- [ ] **Documentation:** Update all references
- [ ] **Client Libraries:** Update with warnings
- [ ] **Support:** Prepare for migration questions

## **🎯 BEST PRACTICES**

### **For API Designers**

1. **Plan for Evolution:** Design APIs to be extensible
2. **Avoid Breaking Changes:** Use optional parameters
3. **Version Early:** Start versioning from v1.0.0
4. **Document Everything:** Clear migration paths
5. **Test Compatibility:** Automated compatibility testing

### **For Client Developers**

1. **Handle Versions:** Check API version in responses
2. **Plan Migrations:** Monitor deprecation warnings
3. **Test Compatibility:** Test with multiple server versions
4. **Update Regularly:** Keep client libraries current
5. **Follow Migration Guides:** Use provided migration tools

## **📚 REFERENCES**

- [Semantic Versioning 2.0.0](https://semver.org/)
- [API Versioning Best Practices](https://restfulapi.net/versioning/)
- [JSON-RPC 2.0 Specification](https://www.jsonrpc.org/specification)
- [WebSocket RFC 6455](https://tools.ietf.org/html/rfc6455)

---

**This versioning policy ensures predictable, stable API evolution while maintaining backward compatibility and providing clear migration paths for clients.**
