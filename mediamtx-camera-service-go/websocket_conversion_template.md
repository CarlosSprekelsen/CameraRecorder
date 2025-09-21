# WebSocket Test Standardization Template

## BEFORE (Verbose Pattern - 12+ lines):
```go
func TestWebSocketMethods_[Method](t *testing.T) {
    helper := NewWebSocketTestHelper(t, nil)
    defer helper.Cleanup(t)
    controller := createMediaMTXControllerUsingProvenPattern(t)
    server := helper.GetServer(t)
    server.SetMediaMTXController(controller)
    server = helper.StartServer(t)
    conn := helper.NewTestClient(t, server)
    defer helper.CleanupTestClient(t, conn)
    AuthenticateTestClient(t, conn, "test_user", "[role]")
    message := CreateTestMessage("[method]", map[string]interface{}{...})
    response := SendTestMessage(t, conn, message)
    // validation...
}
```

## AFTER (Minimal Pattern - 2 lines):
```go
func TestWebSocketMethods_[Method](t *testing.T) {
    helper := NewWebSocketTestHelper(t, nil)
    defer helper.Cleanup(t)
    
    response := helper.TestMethod(t, "[method]", map[string]interface{}{...}, "[role]")
    
    // validation...
}
```

## Conversion Rules:
1. Replace 12-line setup with 2-line setup
2. Use helper.TestMethod() for simple method tests
3. Use helper.GetAuthenticatedConnection() for complex tests
4. Maintain all validation logic
5. Keep all documentation and requirements
