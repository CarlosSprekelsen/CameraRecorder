# Client Project Ground Rules

## Context  
Role-based development with virtual team scaling. Single source of truth for MediaMTX Camera Service Client project.

## Role Authority
- **Developer Role**: Implementation only - cannot claim sprint completion
- **IV&V Role**: Evidence validation - applies quality standards  
- **Project Manager Role**: Sprint completion authority only

## 1. Ground Truth Documents  
Never override without explicit sign-off:
1. **Server API**: `../mediamtx-camera-service/docs/api/json-rpc-methods.md`
2. **Client API**: `docs/api/client-api-reference.md` (must match server)
3. **Testing**: This document (testing rules below)
4. **Architecture**: This document (structure rules below)

## 2. Project Structure
```
MediaMTX-Camera-Service-Client/
├── src/                          # Library (minimal)
├── docs/                         # Documentation
├── tests/                        # Library tests
└── client/                       # React app (MAIN)
    ├── src/                      # React source
    │   ├── types/                # TypeScript definitions
    │   ├── services/             # WebSocket service
    │   └── components/           # React components
    └── package.json              # React dependencies
```

**Rule**: All development happens in `client/` directory. Root level is for docs/tests only.

## 3. Server Alignment Rules
- **Types**: Must match server API exactly
- **Methods**: Use server method names and parameters
- **Errors**: Handle all server error codes
- **Performance**: Meet server targets (<50ms status, <100ms control)

## 4. Testing Rules
- **Real Server**: Use systemd-managed MediaMTX for integration tests
- **No Mocking**: Mock only when server unavailable
- **Coverage**: 80%+ for critical business logic
- **Performance**: Validate against documented targets

## 5. Decision Priority  
1. **Server Implementation** (authoritative)
2. **Existing Ground Truth** (docs above)
3. **Client Requirements** 
4. **Sprint Plan** (Phase 1: S1/S2)
5. **Best Practice** (minimal impact)

## 6. STOP & CLARIFY Policy  
For ambiguous requirements:
```
// STOP: clarify behavior [Story-ID] – specific question here
```
Pause development until answered - no guessing.

## 7. Commit & PR Guidelines
- **One change per PR**: narrow scope
- **Reference story**: `feat(Story-S1): add feature per spec`
- **Link to ground truth**: note doc alignment

## 8. Quality Gates
- **API Compatibility**: All client calls work with server
- **Type Safety**: TypeScript compilation with strict mode
- **Performance**: Meet documented targets
- **Real Integration**: Tests pass against running server

## 9. Evidence Management
- **NO test artifacts in root folder**
- **Use `/evidence/sprint-X/` structure**
- **Archive after sprint completion**

---

**Document Version:** 1.0  
**Status:** Single source of truth for client development
