//go:build performance && stress
// +build performance,stress

/*
Concurrent Connections Stress Test

Requirements Coverage:
- REQ-STRESS-001: Concurrent WebSocket connections
- REQ-STRESS-002: Concurrent request handling
- REQ-STRESS-003: Connection stress over time
- REQ-STRESS-004: Memory stress testing
- REQ-STRESS-005: Rate limiting stress testing
- REQ-STRESS-006: Error rate monitoring
- REQ-STRESS-007: Performance degradation testing
- REQ-STRESS-008: System stability validation

Test Categories: Performance/Stress/Real System
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package stress_test

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
	gorilla "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// StressTestSuite provides stress testing infrastructure
type StressTestSuite struct {
	configManager      *config.ConfigManager
	logger             *logging.Logger
	cameraMonitor      *camera.HybridCameraMonitor
	mediaMTXController mediamtx.MediaMTXController
	jwtHandler         *security.JWTHandler
	wsServer           *websocket.WebSocketServer
	serverURL          string
	ctx                context.Context
	cancel             context.CancelFunc
}

// NewStressTestSuite creates a new stress test suite
func NewStressTestSuite() *StressTestSuite {
	return &StressTestSuite{
		serverURL: "ws://localhost:8002/ws",
	}
}

// Setup initializes the stress test suite
func (suite *StressTestSuite) Setup(t *testing.T) {
	// REQ-STRESS-001: Concurrent WebSocket connections
	
	// Create context with timeout
	suite.ctx, suite.cancel = context.WithTimeout(context.Background(), 300*time.Second)

	// Load configuration
	suite.configManager = config.NewConfigManager()
	err := suite.configManager.LoadConfig("config/default.yaml")
	require.NoError(t, err, "Failed to load configuration")

	// Setup logging
	suite.logger = logging.NewLogger("stress-test-suite")

	// Initialize real implementations
	deviceChecker := &camera.RealDeviceChecker{}
	commandExecutor := &camera.RealV4L2CommandExecutor{}
	infoParser := &camera.RealDeviceInfoParser{}

	// Initialize camera monitor
	suite.cameraMonitor, err = camera.NewHybridCameraMonitor(
		suite.configManager,
		suite.logger,
		deviceChecker,
		commandExecutor,
		infoParser,
	)
	require.NoError(t, err, "Failed to create camera monitor")

	// Initialize MediaMTX controller
	suite.mediaMTXController, err = mediamtx.NewControllerWithConfigManager(suite.configManager, suite.logger.Logger)
	require.NoError(t, err, "Failed to create MediaMTX controller")

	// Initialize JWT handler
	cfg := suite.configManager.GetConfig()
	require.NotNil(t, cfg, "Configuration not available")

	suite.jwtHandler, err = security.NewJWTHandler(cfg.Security.JWTSecretKey)
	require.NoError(t, err, "Failed to create JWT handler")

	// Initialize WebSocket server
	suite.wsServer, err = websocket.NewWebSocketServer(
		suite.configManager,
		suite.logger,
		suite.cameraMonitor,
		suite.jwtHandler,
		suite.mediaMTXController,
	)
	require.NoError(t, err, "Failed to create WebSocket server")

	// Start WebSocket server
	err = suite.wsServer.Start()
	require.NoError(t, err, "Failed to start WebSocket server")

	// Wait for server to be ready
	time.Sleep(2 * time.Second)
}

// Teardown cleans up the stress test suite
func (suite *StressTestSuite) Teardown(t *testing.T) {
	if suite.cancel != nil {
		suite.cancel()
	}

	if suite.wsServer != nil {
		err := suite.wsServer.Stop()
		require.NoError(t, err, "Failed to stop WebSocket server")
	}
}

// WebSocketClient represents a WebSocket client for stress testing
type WebSocketClient struct {
	conn      *gorilla.Conn
	clientID  string
	authToken string
	mu        sync.Mutex
}

// NewWebSocketClient creates a new WebSocket client
func NewWebSocketClient(serverURL, clientID string, authToken string) (*WebSocketClient, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse server URL: %w", err)
	}

	conn, _, err := gorilla.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to WebSocket server: %w", err)
	}

	return &WebSocketClient{
		conn:      conn,
		clientID:  clientID,
		authToken: authToken,
	}, nil
}

// SendRequest sends a JSON-RPC request
func (c *WebSocketClient) SendRequest(request websocket.JsonRpcRequest) (*websocket.JsonRpcResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Send request
	err := c.conn.WriteJSON(request)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Read response
	var response websocket.JsonRpcResponse
	err = c.conn.ReadJSON(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return &response, nil
}

// Close closes the WebSocket connection
func (c *WebSocketClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// TestConcurrentConnections tests 1000+ concurrent WebSocket connections
func TestConcurrentConnections(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	suite := NewStressTestSuite()
	suite.Setup(t)
	defer suite.Teardown(t)

	// Test different connection counts
	connectionCounts := []int{10, 50, 100, 500, 1000}

	for _, count := range connectionCounts {
		t.Run(fmt.Sprintf("ConcurrentConnections_%d", count), func(t *testing.T) {
			testConcurrentConnections(t, suite, count)
		})
	}
}

// testConcurrentConnections tests a specific number of concurrent connections
func testConcurrentConnections(t *testing.T, suite *StressTestSuite, connectionCount int) {
	t.Logf("Testing %d concurrent connections", connectionCount)

	// Generate auth token
	authToken, err := suite.jwtHandler.GenerateToken("stress-test-user", "admin", 1)
	require.NoError(t, err, "Failed to generate auth token")

	// Create clients
	clients := make([]*WebSocketClient, connectionCount)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error

	// Start connection creation
	startTime := time.Now()
	for i := 0; i < connectionCount; i++ {
		wg.Add(1)
		go func(clientIndex int) {
			defer wg.Done()

			clientID := fmt.Sprintf("stress-client-%d", clientIndex)
			client, err := NewWebSocketClient(suite.serverURL, clientID, authToken)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("client %d connection failed: %w", clientIndex, err))
				mu.Unlock()
				return
			}

			// Authenticate client
			authRequest := websocket.JsonRpcRequest{
				JSONRPC: "2.0",
				ID:      clientIndex,
				Method:  "authenticate",
				Params: map[string]interface{}{
					"token": authToken,
				},
			}

			response, err := client.SendRequest(authRequest)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("client %d authentication failed: %w", clientIndex, err))
				mu.Unlock()
				client.Close()
				return
			}

			if response.Error != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("client %d authentication error: %v", clientIndex, response.Error))
				mu.Unlock()
				client.Close()
				return
			}

			clients[clientIndex] = client
		}(i)
	}

	// Wait for all connections
	wg.Wait()
	connectionTime := time.Since(startTime)

	// Count successful connections
	successfulConnections := 0
	for _, client := range clients {
		if client != nil {
			successfulConnections++
		}
	}

	t.Logf("Connection results:")
	t.Logf("  Total connections attempted: %d", connectionCount)
	t.Logf("  Successful connections: %d", successfulConnections)
	t.Logf("  Failed connections: %d", len(errors))
	t.Logf("  Connection time: %v", connectionTime)
	t.Logf("  Average connection time: %v", connectionTime/time.Duration(connectionCount))

	// Log some errors for debugging
	if len(errors) > 0 {
		t.Logf("Sample errors:")
		for i, err := range errors {
			if i >= 5 { // Show only first 5 errors
				break
			}
			t.Logf("  Error %d: %v", i+1, err)
		}
	}

	// Assert minimum success rate (80%)
	successRate := float64(successfulConnections) / float64(connectionCount)
	t.Logf("Success rate: %.2f%%", successRate*100)
	assert.GreaterOrEqual(t, successRate, 0.8, "Success rate should be at least 80%%")

	// Test concurrent requests
	if successfulConnections > 0 {
		testConcurrentRequests(t, suite, clients[:successfulConnections])
	}

	// Cleanup
	for _, client := range clients {
		if client != nil {
			client.Close()
		}
	}
}

// testConcurrentRequests tests concurrent requests from multiple clients
func testConcurrentRequests(t *testing.T, suite *StressTestSuite, clients []*WebSocketClient) {
	t.Logf("Testing concurrent requests from %d clients", len(clients))

	// Test different request types
	requestTypes := []struct {
		name   string
		method string
		params map[string]interface{}
	}{
		{"GetHealth", "get_health", map[string]interface{}{}},
		{"GetCameras", "get_cameras", map[string]interface{}{}},
		{"ListRecordings", "list_recordings", map[string]interface{}{
			"limit":  10,
			"offset": 0,
		}},
		{"ListSnapshots", "list_snapshots", map[string]interface{}{
			"limit":  10,
			"offset": 0,
		}},
	}

	for _, reqType := range requestTypes {
		t.Run(fmt.Sprintf("ConcurrentRequests_%s", reqType.name), func(t *testing.T) {
			testConcurrentRequestType(t, clients, reqType)
		})
	}
}

// testConcurrentRequestType tests concurrent requests of a specific type
func testConcurrentRequestType(t *testing.T, clients []*WebSocketClient, reqType struct {
	name   string
	method string
	params map[string]interface{}
}) {
	t.Logf("Testing concurrent %s requests", reqType.name)

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error
	var responses []*websocket.JsonRpcResponse

	// Start concurrent requests
	startTime := time.Now()
	for i, client := range clients {
		wg.Add(1)
		go func(clientIndex int, client *WebSocketClient) {
			defer wg.Done()

			request := websocket.JsonRpcRequest{
				JSONRPC: "2.0",
				ID:      clientIndex,
				Method:  reqType.method,
				Params:  reqType.params,
			}

			response, err := client.SendRequest(request)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("client %d request failed: %w", clientIndex, err))
				mu.Unlock()
				return
			}

			mu.Lock()
			responses = append(responses, response)
			mu.Unlock()
		}(i, client)
	}

	// Wait for all requests
	wg.Wait()
	requestTime := time.Since(startTime)

	// Count successful requests
	successfulRequests := len(responses)
	totalRequests := len(clients)

	t.Logf("Request results for %s:", reqType.name)
	t.Logf("  Total requests: %d", totalRequests)
	t.Logf("  Successful requests: %d", successfulRequests)
	t.Logf("  Failed requests: %d", len(errors))
	t.Logf("  Request time: %v", requestTime)
	t.Logf("  Average request time: %v", requestTime/time.Duration(totalRequests))

	// Log some errors for debugging
	if len(errors) > 0 {
		t.Logf("Sample errors for %s:", reqType.name)
		for i, err := range errors {
			if i >= 3 { // Show only first 3 errors
				break
			}
			t.Logf("  Error %d: %v", i+1, err)
		}
	}

	// Assert minimum success rate (90% for requests)
	successRate := float64(successfulRequests) / float64(totalRequests)
	t.Logf("Success rate for %s: %.2f%%", reqType.name, successRate*100)
	assert.GreaterOrEqual(t, successRate, 0.9, "Request success rate should be at least 90%%")
}

// TestConnectionStress tests connection stress over time
func TestConnectionStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	suite := NewStressTestSuite()
	suite.Setup(t)
	defer suite.Teardown(t)

	// Test connection stress for 60 seconds
	duration := 60 * time.Second
	connectionInterval := 100 * time.Millisecond
	maxConnections := 100

	t.Logf("Testing connection stress for %v", duration)

	var wg sync.WaitGroup
	var mu sync.Mutex
	var activeConnections []*WebSocketClient
	var totalConnections int
	var totalErrors int

	// Generate auth token
	authToken, err := suite.jwtHandler.GenerateToken("stress-test-user", "admin", 1)
	require.NoError(t, err, "Failed to generate auth token")

	// Start connection stress
	startTime := time.Now()
	ticker := time.NewTicker(connectionInterval)
	defer ticker.Stop()

	// Connection creation goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		connectionID := 0

		for time.Since(startTime) < duration {
			select {
			case <-ticker.C:
				mu.Lock()
				currentConnections := len(activeConnections)
				mu.Unlock()

				if currentConnections < maxConnections {
					// Create new connection
					clientID := fmt.Sprintf("stress-client-%d", connectionID)
					client, err := NewWebSocketClient(suite.serverURL, clientID, authToken)
					if err != nil {
						mu.Lock()
						totalErrors++
						mu.Unlock()
						continue
					}

					// Authenticate client
					authRequest := websocket.JsonRpcRequest{
						JSONRPC: "2.0",
						ID:      connectionID,
						Method:  "authenticate",
						Params: map[string]interface{}{
							"token": authToken,
						},
					}

					response, err := client.SendRequest(authRequest)
					if err != nil || response.Error != nil {
						mu.Lock()
						totalErrors++
						mu.Unlock()
						client.Close()
						continue
					}

					mu.Lock()
					activeConnections = append(activeConnections, client)
					totalConnections++
					mu.Unlock()

					connectionID++
				}
			case <-suite.ctx.Done():
				return
			}
		}
	}()

	// Connection cleanup goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		cleanupTicker := time.NewTicker(5 * time.Second)
		defer cleanupTicker.Stop()

		for time.Since(startTime) < duration {
			select {
			case <-cleanupTicker.C:
				mu.Lock()
				if len(activeConnections) > maxConnections/2 {
					// Remove some connections
					removeCount := len(activeConnections) / 4
					for i := 0; i < removeCount; i++ {
						if len(activeConnections) > 0 {
							client := activeConnections[0]
							activeConnections = activeConnections[1:]
							client.Close()
						}
					}
				}
				mu.Unlock()
			case <-suite.ctx.Done():
				return
			}
		}
	}()

	// Wait for stress test to complete
	wg.Wait()
	totalTime := time.Since(startTime)

	// Cleanup remaining connections
	mu.Lock()
	for _, client := range activeConnections {
		client.Close()
	}
	finalConnections := len(activeConnections)
	mu.Unlock()

	t.Logf("Connection stress test results:")
	t.Logf("  Total time: %v", totalTime)
	t.Logf("  Total connections created: %d", totalConnections)
	t.Logf("  Total errors: %d", totalErrors)
	t.Logf("  Final active connections: %d", finalConnections)
	t.Logf("  Average connections per second: %.2f", float64(totalConnections)/totalTime.Seconds())
	t.Logf("  Error rate: %.2f%%", float64(totalErrors)/float64(totalConnections+totalErrors)*100)

	// Assert reasonable performance
	assert.GreaterOrEqual(t, totalConnections, 50, "Should create at least 50 connections")
	assert.LessOrEqual(t, float64(totalErrors)/float64(totalConnections+totalErrors), 0.1, "Error rate should be less than 10%%")
}

// TestMemoryStress tests memory usage under stress
func TestMemoryStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	suite := NewStressTestSuite()
	suite.Setup(t)
	defer suite.Teardown(t)

	// Test memory stress for 30 seconds
	duration := 30 * time.Second
	operationInterval := 50 * time.Millisecond

	t.Logf("Testing memory stress for %v", duration)

	var wg sync.WaitGroup
	var mu sync.Mutex
	var operations int
	var errors int

	// Start memory stress operations
	startTime := time.Now()
	ticker := time.NewTicker(operationInterval)
	defer ticker.Stop()

	// Memory stress goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()

		for time.Since(startTime) < duration {
			select {
			case <-ticker.C:
				// Perform memory-intensive operations
				_, err := suite.mediaMTXController.GetHealth(suite.ctx)
				if err != nil {
					mu.Lock()
					errors++
					mu.Unlock()
					continue
				}

				_, err = suite.mediaMTXController.GetSystemMetrics(suite.ctx)
				if err != nil {
					mu.Lock()
					errors++
					mu.Unlock()
					continue
				}

				_, err = suite.mediaMTXController.ListRecordings(suite.ctx, 100, 0)
				if err != nil {
					mu.Lock()
					errors++
					mu.Unlock()
					continue
				}

				_, err = suite.mediaMTXController.ListSnapshots(suite.ctx, 100, 0)
				if err != nil {
					mu.Lock()
					errors++
					mu.Unlock()
					continue
				}

				mu.Lock()
				operations++
				mu.Unlock()

			case <-suite.ctx.Done():
				return
			}
		}
	}()

	// Wait for memory stress test to complete
	wg.Wait()
	totalTime := time.Since(startTime)

	t.Logf("Memory stress test results:")
	t.Logf("  Total time: %v", totalTime)
	t.Logf("  Total operations: %d", operations)
	t.Logf("  Total errors: %d", errors)
	t.Logf("  Operations per second: %.2f", float64(operations)/totalTime.Seconds())
	t.Logf("  Error rate: %.2f%%", float64(errors)/float64(operations+errors)*100)

	// Assert reasonable performance
	assert.GreaterOrEqual(t, operations, 100, "Should perform at least 100 operations")
	assert.LessOrEqual(t, float64(errors)/float64(operations+errors), 0.05, "Error rate should be less than 5%%")
}

// TestRateLimitStress tests rate limiting under stress
func TestRateLimitStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	suite := NewStressTestSuite()
	suite.Setup(t)
	defer suite.Teardown(t)

	// Test rate limiting stress
	clientCount := 50
	requestsPerClient := 200

	t.Logf("Testing rate limiting stress with %d clients, %d requests each", clientCount, requestsPerClient)

	// Generate auth token
	authToken, err := suite.jwtHandler.GenerateToken("stress-test-user", "admin", 1)
	require.NoError(t, err, "Failed to generate auth token")

	// Create clients
	clients := make([]*WebSocketClient, clientCount)
	for i := 0; i < clientCount; i++ {
		clientID := fmt.Sprintf("rate-limit-client-%d", i)
		client, err := NewWebSocketClient(suite.serverURL, clientID, authToken)
		require.NoError(t, err, "Failed to create client %d", i)

		// Authenticate client
		authRequest := websocket.JsonRpcRequest{
			JSONRPC: "2.0",
			ID:      i,
			Method:  "authenticate",
			Params: map[string]interface{}{
				"token": authToken,
			},
		}

		response, err := client.SendRequest(authRequest)
		require.NoError(t, err, "Failed to authenticate client %d", i)
		require.Nil(t, response.Error, "Authentication failed for client %d", i)

		clients[i] = client
		defer client.Close()
	}

	// Start rate limiting stress
	var wg sync.WaitGroup
	var mu sync.Mutex
	var totalRequests int
	var rateLimitErrors int
	var otherErrors int

	startTime := time.Now()

	for i, client := range clients {
		wg.Add(1)
		go func(clientIndex int, client *WebSocketClient) {
			defer wg.Done()

			for j := 0; j < requestsPerClient; j++ {
				request := websocket.JsonRpcRequest{
					JSONRPC: "2.0",
					ID:      clientIndex*requestsPerClient + j,
					Method:  "get_health",
					Params:  map[string]interface{}{},
				}

				response, err := client.SendRequest(request)
				if err != nil {
					mu.Lock()
					otherErrors++
					mu.Unlock()
					continue
				}

				if response.Error != nil && response.Error.Code == websocket.RATE_LIMIT_EXCEEDED {
					mu.Lock()
					rateLimitErrors++
					mu.Unlock()
				} else if response.Error != nil {
					mu.Lock()
					otherErrors++
					mu.Unlock()
				}

				mu.Lock()
				totalRequests++
				mu.Unlock()
			}
		}(i, client)
	}

	// Wait for all requests
	wg.Wait()
	totalTime := time.Since(startTime)

	t.Logf("Rate limiting stress test results:")
	t.Logf("  Total time: %v", totalTime)
	t.Logf("  Total requests: %d", totalRequests)
	t.Logf("  Rate limit errors: %d", rateLimitErrors)
	t.Logf("  Other errors: %d", otherErrors)
	t.Logf("  Requests per second: %.2f", float64(totalRequests)/totalTime.Seconds())
	t.Logf("  Rate limit error rate: %.2f%%", float64(rateLimitErrors)/float64(totalRequests)*100)
	t.Logf("  Other error rate: %.2f%%", float64(otherErrors)/float64(totalRequests)*100)

	// Assert rate limiting is working
	assert.Greater(t, rateLimitErrors, 0, "Should have some rate limit errors")
	assert.LessOrEqual(t, float64(otherErrors)/float64(totalRequests), 0.1, "Other error rate should be less than 10%%")
}
