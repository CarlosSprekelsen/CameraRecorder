// Package config implements ConfigStore from Architecture §5.
package config

import "time"

// TimingConfig maps CB-TIMING v0.3 structure.
type TimingConfig struct {
	// CB-TIMING §3.1 Heartbeat Configuration
	HeartbeatInterval time.Duration
	HeartbeatJitter   time.Duration
	HeartbeatTimeout  time.Duration

	// CB-TIMING §4.1 Probe States & Cadences
	ProbeNormalInterval    time.Duration
	ProbeRecoveringInitial time.Duration
	ProbeRecoveringBackoff float64
	ProbeRecoveringMax     time.Duration
	ProbeOfflineInitial    time.Duration
	ProbeOfflineBackoff    float64
	ProbeOfflineMax        time.Duration

	// CB-TIMING §5 Command Timeout Classes
	CommandTimeoutSetPower    time.Duration
	CommandTimeoutSetChannel  time.Duration
	CommandTimeoutSelectRadio time.Duration
	CommandTimeoutGetState    time.Duration

	// CB-TIMING §6.1 Event Buffer Configuration
	EventBufferSize      int
	EventBufferRetention time.Duration
}

// LoadCBTimingBaseline returns CB-TIMING v0.3 baseline values.
func LoadCBTimingBaseline() *TimingConfig {
	return &TimingConfig{
		// CB-TIMING §3.1: Heartbeat interval 15s, jitter ±2s, timeout 45s
		HeartbeatInterval: 15 * time.Second, // CB-TIMING §3.1
		HeartbeatJitter:   2 * time.Second,  // CB-TIMING §3.1
		HeartbeatTimeout:  45 * time.Second, // CB-TIMING §3.1

		// CB-TIMING §4.1: Normal 30s, Recovering 5s/1.5x/15s, Offline 10s/2.0x/300s
		ProbeNormalInterval:    30 * time.Second,  // CB-TIMING §4.1
		ProbeRecoveringInitial: 5 * time.Second,   // CB-TIMING §4.1
		ProbeRecoveringBackoff: 1.5,               // CB-TIMING §4.1
		ProbeRecoveringMax:     15 * time.Second,  // CB-TIMING §4.1
		ProbeOfflineInitial:    10 * time.Second,  // CB-TIMING §4.1
		ProbeOfflineBackoff:    2.0,               // CB-TIMING §4.1
		ProbeOfflineMax:        300 * time.Second, // CB-TIMING §4.1

		// CB-TIMING §5: setPower 10s, setChannel 30s, selectRadio 5s, getState 5s
		CommandTimeoutSetPower:    10 * time.Second, // CB-TIMING §5
		CommandTimeoutSetChannel:  30 * time.Second, // CB-TIMING §5
		CommandTimeoutSelectRadio: 5 * time.Second,  // CB-TIMING §5
		CommandTimeoutGetState:    5 * time.Second,  // CB-TIMING §5

		// CB-TIMING §6.1: 50 events, 1 hour retention
		EventBufferSize:      50,            // CB-TIMING §6.1
		EventBufferRetention: 1 * time.Hour, // CB-TIMING §6.1
	}
}
