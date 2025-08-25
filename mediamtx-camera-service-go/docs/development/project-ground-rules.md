# Project Ground Rules

## 🚨 **CRITICAL: STOP and Authorization Required**
- **STOP before modifying any code** - Investigate and understand the issue first
- **Ask for explicit authorization** before making any code changes
- **No working in isolation** - Coordinate with team before implementation
- **Present options and recommendations** for team decision
- **Do not create document over population** - Only create requested reports, do not offer free reports outside the chat unless requested

## 🚨 **CRITICAL: API Documentation is Ground Truth**
- **API Documentation**: `docs/api/json_rpc_methods.md` is the ONLY source of truth for API behavior
- **Health Endpoints**: `docs/api/health-endpoints.md` is the ONLY source of truth for health API
- **NEVER use server implementation as reference** - Only use documented API
- **Tests must validate against API documentation** - Not against server implementation
- **If test fails, check API documentation first** - Don't adapt test to broken implementation

## **Ground Truth Enforcement Rules**
1. **API Documentation is FROZEN** - Changes require formal approval process
2. **Server Implementation follows API Documentation** - Not the other way around
3. **Tests validate API compliance** - Not implementation details
4. **Test failures indicate API/implementation mismatch** - Not test bugs
5. **No "accommodation" of broken implementations** - Tests do not fix the implementation - it is ok if a test fails, that's their purpose, to find real bugs not accommodate them

## **Real System Testing Over Mocking**
- **MediaMTX:** Use systemd-managed service, never mock
- **File System:** Use `tempfile`, never mock
- **WebSocket:** Use real connections within system
- **Authentication:** Use real JWT tokens with test secrets
- **API Keys:** Use test-accessible storage location (`/tmp/test_api_keys.json`)

## **Strategic Mocking Rules**
**MOCK:** External APIs, time operations, expensive hardware simulation  
**NEVER MOCK:** MediaMTX service, filesystem, internal WebSocket, JWT auth, config loading

---

## Context  
Role-based development with virtual team scaling. These ground rules define single source of truth, decision priorities, and role boundaries for any project.

## Role Authority
- **Developer Role**: Implementation only - cannot claim sprint completion
- **IV&V Role**: Evidence validation - applies quality standards from roles document  
- **Project Manager Role**: Sprint completion authority only
- **Role reference**: See roles document in project knowledge for specific boundaries

---

## 1. Ground Truth Documents  
Never override without explicit sign-off:
1. **Architecture Overview**: `docs/architecture/overview.md` (project-specific)
2. **Development Principles**: `docs/development/principles.md` (universal)
3. **Documentation Guidelines**: `docs/development/documentation-guidelines.md` (universal)  
4. **Roles and Responsibilities**: `docs/development/roles-responsibilities.md` (universal)
5. **API Reference**: `docs/api/` (project-specific)
6. **Project Requirements**: Project requirements document (project-specific)
7. **Test Guidelines**: `docs/development/testing-guide.md` (project-specific)

## 2. Decision Priority  
When encountering conflicting information:
1. **Existing Ground Truth**: Follow docs above (including roles authority)
2. **Client Requirements**: Apply `client-requirements.md` rules  
3. **Roadmap Stories**: Adhere to sprint plan (Phase 1: S1/S2)  
4. **Best Practice**: Minimal impact following principles guidelines

## 3. Scope & "No Scope Creep"  
- **MVP = Phase 1** only: features defined in current sprint/phase requirements
- **Everything else → Phase 2+**: features beyond current scope  
- **Any "nice-to-have"** outside MVP goes to backlog with "Phase X" label

## 4. STOP & CLARIFY Policy  
For ambiguous requirements:
1. **Insert canonical STOP comment** at decision point:
   ```
   // STOP: clarify behavior for edge case [Story-ID] – Should system retry or fail immediately?
   ```
2. **Raise precise question** referencing STOP tag and story
3. **Pause development** until answered - no guessing

## 5. Commit & PR Guidelines
- **One change per PR**: narrow scope
- **Reference story & STOP tags**: `feat(Story-S1): add feature per spec`
- **Link to ground truth**: note doc alignment in commit message

## 6. Testing & Validation
- **"Green bar"**: all tests + lint + type-check pass
- **Stop on first failure**: fix before writing new tests
- **Test naming**: map to stories (`test_feature_story_S1`)
- **No silent skips**: TODO tests must `throw new Error("STOP: implement...")`

## 7. Evidence Management (Universal Pattern)
- **NO test artifacts in root folder**
- **Use `/evidence/sprint-X/` structure for role coordination**
- **Evidence = communication tool between Developer → IV&V → PM**
- **Archive or remove after sprint completion to keep code clean**
