//go:build integration

package command

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/internal/command"
	"github.com/radio-control/rcc/internal/config"
	"github.com/radio-control/rcc/internal/telemetry"
	"github.com/radio-control/rcc/test/integration/fakes"
)

// MockAuditLogger captures audit calls for verification
type MockAuditLogger struct {
	LoggedActions []AuditCall
}

type AuditCall struct {
	Action  string
	RadioID string
	Result  string
	Latency time.Duration
}

func (m *MockAuditLogger) LogAction(ctx context.Context, action, radioID, result string, latency time.Duration) {
	m.LoggedActions = append(m.LoggedActions, AuditCall{
		Action:  action,
		RadioID: radioID,
		Result:  result,
		Latency: latency,
	})
}

// MockRadioManager provides radio management for testing
type MockRadioManager struct {
	activeRadioID string
	radios        map[string]*adapter.RadioCapabilities
}

func NewMockRadioManager() *MockRadioManager {
	return &MockRadioManager{
		radios: make(map[string]*adapter.RadioCapabilities),
	}
}

func (m *MockRadioManager) LoadCapabilities(radioID string, adapter adapter.IRadioAdapter, timeout time.Duration) error {
	// Extract capabilities from adapter using SupportedFrequencyProfiles
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	profiles, err := adapter.SupportedFrequencyProfiles(ctx)
	if err != nil {
		return err
	}
	
	// Convert frequency profiles to channels
	channels := make([]adapter.Channel, 0)
	for i, profile := range profiles {
		for j, freq := range profile.Frequencies {
			channels = append(channels, adapter.Channel{
				Index:        i*len(profile.Frequencies) + j + 1, // 1-based indexing
				FrequencyMhz: freq,
			})
		}
	}
	
	m.radios[radioID] = &adapter.RadioCapabilities{
		Channels: channels,
	}
	m.activeRadioID = radioID
	return nil
}

func (m *MockRadioManager) GetActiveRadioID() string {
	return m.activeRadioID
}

func (m *MockRadioManager) GetChannelByIndex(radioID string, index int) (adapter.Channel, error) {
	caps, exists := m.radios[radioID]
	if !exists {
		return adapter.Channel{}, adapter.ErrUnavailable
	}
	
	for _, ch := range caps.Channels {
		if ch.Index == index {
			return ch, nil
		}
	}
	
	return adapter.Channel{}, adapter.ErrUnavailable
}

// TestCommand_SetPower_IntegrationFlow tests the complete command flow through interfaces
func TestCommand_SetPower_IntegrationFlow(t *testing.T) {
	// Arrange: Setup mocks and real components
	mockAudit := &MockAuditLogger{}
	mockRadioManager := NewMockRadioManager()
	
	// Create fake adapter with known state
	fakeAdapter := fakes.NewFakeAdapter("test-radio-01").
		WithInitial(20.0, 2412.0, []adapter.Channel{
			{Index: 1, FrequencyMhz: 2412.0},
			{Index: 6, FrequencyMhz: 2437.0},
			{Index: 11, FrequencyMhz: 2462.0},
		})
	
	// Load capabilities into radio manager
	err := mockRadioManager.LoadCapabilities("test-radio-01", fakeAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load capabilities: %v", err)
	}
	
	// Create real telemetry hub for integration testing
	telemetryHub := telemetry.NewHub(&config.TimingConfig{})
	
	// Create orchestrator with real telemetry hub
	orchestrator := command.NewOrchestratorWithRadioManager(telemetryHub, &config.TimingConfig{}, mockRadioManager)
	orchestrator.SetAuditLogger(mockAudit)
	orchestrator.SetActiveAdapter(fakeAdapter)
	
	// Act: Execute SetPower command
	ctx := context.Background()
	err = orchestrator.SetPower(ctx, "test-radio-01", 25.0)
	
	// Assert: Command execution
	if err != nil {
		t.Errorf("SetPower failed: %v", err)
	}
	
	// Assert: Audit logging
	if len(mockAudit.LoggedActions) == 0 {
		t.Error("Expected audit log entry, but none was recorded")
	} else {
		action := mockAudit.LoggedActions[0]
		if action.Action != "setPower" {
			t.Errorf("Expected action 'setPower', got '%s'", action.Action)
		}
		if action.RadioID != "test-radio-01" {
			t.Errorf("Expected radioID 'test-radio-01', got '%s'", action.RadioID)
		}
		if action.Result != "SUCCESS" {
			t.Errorf("Expected result 'SUCCESS', got '%s'", action.Result)
		}
	}
	
	// Assert: Adapter was called
	if fakeAdapter.GetCallCount("SetPower") != 1 {
		t.Errorf("Expected SetPower to be called once, got %d calls", fakeAdapter.GetCallCount("SetPower"))
	}
	if fakeAdapter.GetLastSetPowerCall() != 25.0 {
		t.Errorf("Expected SetPower(25.0), got SetPower(%f)", fakeAdapter.GetLastSetPowerCall())
	}
	
	// Note: Telemetry events are published to the real hub
	// For integration testing, we verify the command flow executes successfully
	// Telemetry event validation is handled in unit tests
	
	t.Logf("✅ SetPower integration flow: Command → Audit → Adapter")
}

// TestCommand_SetChannelByIndex_IntegrationFlow tests channel index resolution
func TestCommand_SetChannelByIndex_IntegrationFlow(t *testing.T) {
	// Arrange: Setup mocks and real components
	mockAudit := &MockAuditLogger{}
	mockRadioManager := NewMockRadioManager()
	
	// Create fake adapter with known band plan
	fakeAdapter := fakes.NewFakeAdapter("test-radio-01").
		WithInitial(20.0, 2412.0, []adapter.Channel{
			{Index: 1, FrequencyMhz: 2412.0},
			{Index: 6, FrequencyMhz: 2437.0},
			{Index: 11, FrequencyMhz: 2462.0},
		})
	
	// Load capabilities into radio manager
	err := mockRadioManager.LoadCapabilities("test-radio-01", fakeAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load capabilities: %v", err)
	}
	
	// Create real telemetry hub for integration testing
	telemetryHub := telemetry.NewHub(&config.TimingConfig{})
	
	// Create orchestrator with real telemetry hub
	orchestrator := command.NewOrchestratorWithRadioManager(telemetryHub, &config.TimingConfig{}, mockRadioManager)
	orchestrator.SetAuditLogger(mockAudit)
	orchestrator.SetActiveAdapter(fakeAdapter)
	
	// Act: Execute SetChannelByIndex command
	ctx := context.Background()
	err = orchestrator.SetChannelByIndex(ctx, "test-radio-01", 6, mockRadioManager)
	
	// Assert: Command execution should succeed
	if err != nil {
		t.Errorf("SetChannelByIndex failed: %v", err)
	}
	
	// Assert: Adapter was called with correct frequency
	if fakeAdapter.GetCallCount("SetFrequency") != 1 {
		t.Errorf("Expected SetFrequency to be called once, got %d calls", fakeAdapter.GetCallCount("SetFrequency"))
	}
	if fakeAdapter.GetLastSetFrequencyCall() != 2437.0 {
		t.Errorf("Expected SetFrequency(2437.0), got SetFrequency(%f)", fakeAdapter.GetLastSetFrequencyCall())
	}
	
	// Assert: Audit logging
	if len(mockAudit.LoggedActions) == 0 {
		t.Error("Expected audit log entry, but none was recorded")
	} else {
		action := mockAudit.LoggedActions[0]
		if action.Action != "setChannel" {
			t.Errorf("Expected action 'setChannel', got '%s'", action.Action)
		}
		if action.RadioID != "test-radio-01" {
			t.Errorf("Expected radioID 'test-radio-01', got '%s'", action.RadioID)
		}
		if action.Result != "SUCCESS" {
			t.Errorf("Expected result 'SUCCESS', got '%s'", action.Result)
		}
	}
	
	t.Logf("✅ SetChannelByIndex integration flow: Index 6 → Frequency 2437.0 → Adapter")
}

// TestCommand_ErrorNormalization_IntegrationFlow tests error handling through interfaces
func TestCommand_ErrorNormalization_IntegrationFlow(t *testing.T) {
	testCases := []struct {
		name        string
		mode        string
		operation   string
		expectedErr string
	}{
		{
			name:        "Happy mode - SetPower",
			mode:        "happy",
			operation:   "SetPower",
			expectedErr: "",
		},
		{
			name:        "Busy mode - SetPower",
			mode:        "busy",
			operation:   "SetPower",
			expectedErr: "BUSY",
		},
		{
			name:        "Unavailable mode - SetPower",
			mode:        "unavailable",
			operation:   "SetPower",
			expectedErr: "UNAVAILABLE",
		},
		{
			name:        "Invalid range mode - SetPower",
			mode:        "invalid-range",
			operation:   "SetPower",
			expectedErr: "INVALID_RANGE",
		},
		{
			name:        "Internal mode - SetPower",
			mode:        "internal",
			operation:   "SetPower",
			expectedErr: "INTERNAL",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange: Setup mocks and real components
			mockAudit := &MockAuditLogger{}
			mockRadioManager := NewMockRadioManager()
			
			// Create fake adapter in specific mode
			fakeAdapter := fakes.NewFakeAdapter("test-radio-01").
				WithMode(tc.mode).
				WithInitial(20.0, 2412.0, []adapter.Channel{
					{Index: 1, FrequencyMhz: 2412.0},
				})
			
			// Load capabilities into radio manager
			err := mockRadioManager.LoadCapabilities("test-radio-01", fakeAdapter, 5*time.Second)
			if err != nil {
				t.Fatalf("Failed to load capabilities: %v", err)
			}
			
			// Create real telemetry hub for integration testing
			telemetryHub := telemetry.NewHub(&config.TimingConfig{})
			
			// Create orchestrator with real telemetry hub
			orchestrator := command.NewOrchestratorWithRadioManager(telemetryHub, &config.TimingConfig{}, mockRadioManager)
			orchestrator.SetAuditLogger(mockAudit)
			orchestrator.SetActiveAdapter(fakeAdapter)
			
			// Act: Execute operation
			ctx := context.Background()
			var operationErr error
			
			switch tc.operation {
			case "SetPower":
				operationErr = orchestrator.SetPower(ctx, "test-radio-01", 25.0)
			default:
				t.Fatalf("Unknown operation: %s", tc.operation)
			}
			
			// Assert: Error normalization
			if tc.expectedErr == "" {
				if operationErr != nil {
					t.Errorf("Expected success, but got error: %v", operationErr)
				}
				// Success should emit audit
				if len(mockAudit.LoggedActions) == 0 {
					t.Error("Expected audit log entry for successful operation")
				}
			} else {
				if operationErr == nil {
					t.Errorf("Expected %s error, but got none", tc.expectedErr)
				} else {
					if !strings.Contains(operationErr.Error(), tc.expectedErr) {
						t.Errorf("Expected error to contain '%s', got: %v", tc.expectedErr, operationErr)
					}
				}
			}
			
			t.Logf("✅ %s: %s", tc.operation, tc.name)
		})
	}
}