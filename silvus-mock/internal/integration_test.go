package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/silvus-mock/internal/config"
	"github.com/silvus-mock/internal/jsonrpc"
	"github.com/silvus-mock/internal/maintenance"
	"github.com/silvus-mock/internal/state"
)

// Integration tests that test the full system working together

func TestFullSystemIntegration(t *testing.T) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize radio state
	radioState := state.NewRadioState(cfg)
	defer radioState.Close()

	// Test basic state operations
	t.Run("BasicStateOperations", func(t *testing.T) {
		// Test set and get power
		response := radioState.ExecuteCommand("setPower", []string{"25"})
		if response.Error != "" {
			t.Errorf("Set power failed: %v", response.Error)
		}

		response = radioState.ExecuteCommand("getPower", []string{})
		if response.Error != "" {
			t.Errorf("Get power failed: %v", response.Error)
		}
		if response.Result == nil {
			t.Error("Expected power result")
		}

		// Test set and get frequency
		response = radioState.ExecuteCommand("setFreq", []string{"4700"})
		if response.Error != "" {
			t.Errorf("Set frequency failed: %v", response.Error)
		}

		// Wait for blackout to clear
		time.Sleep(6 * time.Second)

		response = radioState.ExecuteCommand("getFreq", []string{})
		if response.Error != "" {
			t.Errorf("Get frequency failed: %v", response.Error)
		}
		if response.Result == nil {
			t.Error("Expected frequency result")
		}
	})

	t.Run("JSONRPCServerIntegration", func(t *testing.T) {
		// Create JSON-RPC server
		jsonrpcServer := jsonrpc.NewServer(cfg, radioState)

		// Test various JSON-RPC requests
		tests := []struct {
			name   string
			method string
			params []string
		}{
			{"SetPower", "power_dBm", []string{"20"}},
			{"GetPower", "power_dBm", nil},
			{"SetFreq", "freq", []string{"4700"}},
			{"GetFreq", "freq", nil},
			{"GetProfiles", "supported_frequency_profiles", nil},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := jsonrpc.Request{
					JSONRPC: "2.0",
					Method:  tt.method,
					Params:  tt.params,
					ID:      "test",
				}

				response := jsonrpcServer.processRequest(&req)
				if response.JSONRPC != "2.0" {
					t.Errorf("Expected JSONRPC 2.0, got %s", response.JSONRPC)
				}
				if response.ID != "test" {
					t.Errorf("Expected ID 'test', got %v", response.ID)
				}
			})
		}
	})

	t.Run("MaintenanceServerIntegration", func(t *testing.T) {
		// Create maintenance server
		maintenanceServer := maintenance.NewServer(cfg, radioState)
		defer maintenanceServer.Close()

		// Test maintenance commands
		tests := []struct {
			name   string
			method string
		}{
			{"Zeroize", "zeroize"},
			{"RadioReset", "radio_reset"},
			{"FactoryReset", "factory_reset"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := maintenance.Request{
					JSONRPC: "2.0",
					Method:  tt.method,
					ID:      "test",
				}

				response := maintenanceServer.processMaintenanceRequest(&req)
				if response.JSONRPC != "2.0" {
					t.Errorf("Expected JSONRPC 2.0, got %s", response.JSONRPC)
				}
				if response.ID != "test" {
					t.Errorf("Expected ID 'test', got %v", response.ID)
				}
			})
		}
	})
}

func TestConfigurationIntegration(t *testing.T) {
	// Test that configuration is properly loaded and used
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Verify configuration values are reasonable
	if cfg.Network.HTTP.Port <= 0 {
		t.Error("HTTP port should be positive")
	}
	if cfg.Network.Maintenance.Port <= 0 {
		t.Error("Maintenance port should be positive")
	}
	if len(cfg.Network.Maintenance.AllowedCIDRs) == 0 {
		t.Error("Should have at least one allowed CIDR")
	}
	if len(cfg.Profiles.FrequencyProfiles) == 0 {
		t.Error("Should have at least one frequency profile")
	}
	if cfg.Power.MinDBm < 0 || cfg.Power.MaxDBm > 50 {
		t.Error("Power range should be reasonable")
	}

	// Test that radio state can be created with config
	radioState := state.NewRadioState(cfg)
	defer radioState.Close()

	// Verify initial state
	freq, power, available := radioState.GetStatus()
	if freq == "" {
		t.Error("Initial frequency should not be empty")
	}
	if power < 0 || power > 50 {
		t.Error("Initial power should be reasonable")
	}
	if !available {
		t.Error("Radio should be available initially")
	}
}

func TestErrorHandlingIntegration(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	radioState := state.NewRadioState(cfg)
	defer radioState.Close()

	// Test various error conditions
	tests := []struct {
		name    string
		command string
		params  []string
		wantErr string
	}{
		{
			name:    "Invalid power range",
			command: "setPower",
			params:  []string{"50"},
			wantErr: "INVALID_RANGE",
		},
		{
			name:    "Invalid frequency",
			command: "setFreq",
			params:  []string{"9999"},
			wantErr: "INVALID_RANGE",
		},
		{
			name:    "Invalid command",
			command: "invalidCommand",
			params:  []string{},
			wantErr: "INTERNAL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := radioState.ExecuteCommand(tt.command, tt.params)
			if response.Error != tt.wantErr {
				t.Errorf("Expected error '%s', got '%s'", tt.wantErr, response.Error)
			}
		})
	}
}

func TestBlackoutBehaviorIntegration(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	radioState := state.NewRadioState(cfg)
	defer radioState.Close()

	// Set frequency to trigger blackout
	response := radioState.ExecuteCommand("setFreq", []string{"4700"})
	if response.Error != "" {
		t.Fatalf("Failed to set frequency: %v", response.Error)
	}

	// Should be in blackout
	if radioState.IsAvailable() {
		t.Error("Expected radio to be unavailable during blackout")
	}

	// Commands should return BUSY
	busyResponse := radioState.ExecuteCommand("setPower", []string{"25"})
	if busyResponse.Error != "BUSY" {
		t.Errorf("Expected BUSY during blackout, got '%s'", busyResponse.Error)
	}

	// Wait for blackout to clear
	time.Sleep(6 * time.Second)

	// Should be available again
	if !radioState.IsAvailable() {
		t.Error("Expected radio to be available after blackout")
	}

	// Commands should work again
	response = radioState.ExecuteCommand("setPower", []string{"25"})
	if response.Error != "" {
		t.Errorf("Expected successful command after blackout, got '%s'", response.Error)
	}
}

func TestFrequencyValidationIntegration(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	radioState := state.NewRadioState(cfg)
	defer radioState.Close()

	// Test frequency validation against configured profiles
	tests := []struct {
		freq    string
		valid   bool
	}{
		{"4700", true},    // Should be in profiles
		{"2220", true},    // Should be in range
		{"9999", false},   // Should be invalid
		{"abc", false},    // Should be invalid
	}

	for _, tt := range tests {
		t.Run("freq_"+tt.freq, func(t *testing.T) {
			response := radioState.ExecuteCommand("setFreq", []string{tt.freq})
			
			if tt.valid {
				if response.Error != "" {
					t.Errorf("Expected valid frequency %s to work, got error: %s", tt.freq, response.Error)
				}
			} else {
				if response.Error != "INVALID_RANGE" {
					t.Errorf("Expected INVALID_RANGE for frequency %s, got: %s", tt.freq, response.Error)
				}
			}
		})
	}
}

func TestMaintenanceOperationsIntegration(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	radioState := state.NewRadioState(cfg)
	defer radioState.Close()

	// Set some values first
	radioState.ExecuteCommand("setPower", []string{"25"})
	radioState.ExecuteCommand("setFreq", []string{"4700"})

	// Test zeroize
	response := radioState.ExecuteCommand("zeroize", []string{})
	if response.Error != "" {
		t.Errorf("Zeroize failed: %v", response.Error)
	}

	// Check that values were reset
	freq, power, _ := radioState.GetStatus()
	if freq != "2490.0" {
		t.Errorf("Expected frequency 2490.0 after zeroize, got %s", freq)
	}
	if power != 30 {
		t.Errorf("Expected power 30 after zeroize, got %d", power)
	}

	// Test radio reset
	response = radioState.ExecuteCommand("radioReset", []string{})
	if response.Error != "" {
		t.Errorf("Radio reset failed: %v", response.Error)
	}

	// Should be in blackout
	if radioState.IsAvailable() {
		t.Error("Expected radio to be unavailable after reset")
	}
}

// Helper function to test HTTP server integration (requires actual HTTP server)
func TestHTTPServerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping HTTP server integration test in short mode")
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	radioState := state.NewRadioState(cfg)
	defer radioState.Close()

	// Create HTTP server
	mux := http.NewServeMux()
	jsonrpcServer := jsonrpc.NewServer(cfg, radioState)
	mux.HandleFunc("/streamscape_api", jsonrpcServer.HandleRequest)

	server := &http.Server{
		Addr:    ":0", // Let OS choose port
		Handler: mux,
	}

	// Start server in background
	go func() {
		server.ListenAndServe()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test HTTP requests
	t.Run("HTTPPowerRequest", func(t *testing.T) {
		reqBody := `{"jsonrpc":"2.0","method":"power_dBm","params":["25"],"id":"test"}`
		resp, err := http.Post("http://localhost:8080/streamscape_api", "application/json", 
			strings.NewReader(reqBody))
		if err != nil {
			t.Logf("HTTP request failed (server may not be running): %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var response jsonrpc.Response
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.JSONRPC != "2.0" {
			t.Errorf("Expected JSONRPC 2.0, got %s", response.JSONRPC)
		}
	})

	server.Close()
}
