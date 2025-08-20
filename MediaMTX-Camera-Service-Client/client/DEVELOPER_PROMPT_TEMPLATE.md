## 🚨 CRITICAL CONTEXT & CONSTRAINTS

### **MANDATORY TEST GUIDELINES (NON-NEGOTIABLE): defined in `client/docs/development/client-testing-guidelines.md`**
1. **"Real Integration Always"** - NO MOCKS in integration tests
2. **Environment Setup Required** - ALWAYS run `source .test_env` before tests
3. **Authentication Setup Required** - ALWAYS run `./set-test-env.sh` before tests
4. **Correct Endpoints** - Use port 8002 for WebSocket, port 8003 for Health
5. **Stable Fixtures** - Use `WebSocketTestFixture` and `HealthTestFixture` exclusively
6. **Jest Patterns** - Use `describe`/`test` structure, NOT function-based tests

### **ARCHITECTURE CONSTRAINTS:**
- **Integration Tests:** Run in Node.js environment (NO DOM)
- **Unit Tests:** Run in jsdom environment (React components)
- **WebSocket Server:** Port 8002 (JSON-RPC methods)
- **Health Server:** Port 8003 (REST endpoints)
- **Authentication:** JWT tokens via `CAMERA_SERVICE_JWT_SECRET`

### **COMMON FAILURES TO AVOID:**
- ❌ **React DOM in Node.js** - `renderHook` causes "document is not defined"
- ❌ **Mixed Architecture** - Don't combine stable fixtures with custom services
- ❌ **Function-based Tests** - Use Jest `describe`/`test` structure
- ❌ **Manual Result Tracking** - Use Jest assertions, not manual counters
- ❌ **Over-Engineering** - Focus on real integration, not complex validation
- ❌ **Forcing Tests to Pass** - Fix design issues, don't force false positives


## 🎯 TASK: Fix Issue [ISSUE_NAME]

## 🚨 CRITICAL REMINDERS

### **DO NOT:**
- ❌ Force tests to pass with false positives
- ❌ Add mocks in integration tests
- ❌ Skip environment setup
- ❌ Use React DOM in Node.js environment
- ❌ Mix stable fixtures with custom services
- ❌ Use function-based tests instead of Jest structure
- ❌ Track results manually instead of using assertions

### **DO:**
- ✅ Fix underlying design issues
- ✅ Use stable fixtures exclusively
- ✅ Follow Jest patterns properly
- ✅ Test real server communication
- ✅ Validate actual requirements
- ✅ Maintain test isolation and cleanup
- ✅ Use proper assertions for each requirement

### **CONTEXT:**
- **Server is working correctly** - Issues are client-side test design
- **Requirements are valid** - Tests correctly validate real requirements
- **Focus on architecture** - Fix design patterns, not requirements
- **Maintain consistency** - Use same patterns as working tests


## 📋 CHECKLIST

Before submitting your fix, ensure:

- [ ] No React DOM dependencies in integration tests
- [ ] Uses stable fixtures exclusively
- [ ] Follows Jest describe/test structure
- [ ] Has proper assertions for each requirement
- [ ] Tests run without environment errors
- [ ] Maintains requirement coverage
- [ ] Uses correct endpoints (8002/8003)
- [ ] Proper authentication setup
- [ ] No manual result tracking
- [ ] Simplified architecture focused on real integration


**Remember: The goal is to fix the test design to properly validate requirements, NOT to force tests to pass. The server API is working correctly - the issues are in the test architecture.**
