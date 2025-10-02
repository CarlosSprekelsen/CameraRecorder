package config

import (
	"testing"
	"time"
)

func TestLoadCBTimingBaseline(t *testing.T) {
	cfg := LoadCBTimingBaseline()

	// CB-TIMING §3.1
	if cfg.HeartbeatInterval != 15*time.Second {
		t.Errorf("HeartbeatInterval = %v, want 15s", cfg.HeartbeatInterval)
	}
	if cfg.HeartbeatJitter != 2*time.Second {
		t.Errorf("HeartbeatJitter = %v, want 2s", cfg.HeartbeatJitter)
	}
	if cfg.HeartbeatTimeout != 45*time.Second {
		t.Errorf("HeartbeatTimeout = %v, want 45s", cfg.HeartbeatTimeout)
	}

	// CB-TIMING §4.1
	if cfg.ProbeNormalInterval != 30*time.Second {
		t.Errorf("ProbeNormalInterval = %v, want 30s", cfg.ProbeNormalInterval)
	}
	if cfg.ProbeRecoveringInitial != 5*time.Second {
		t.Errorf("ProbeRecoveringInitial = %v, want 5s", cfg.ProbeRecoveringInitial)
	}
	if cfg.ProbeRecoveringBackoff != 1.5 {
		t.Errorf("ProbeRecoveringBackoff = %v, want 1.5", cfg.ProbeRecoveringBackoff)
	}

	// CB-TIMING §5
	if cfg.CommandTimeoutSetPower != 10*time.Second {
		t.Errorf("CommandTimeoutSetPower = %v, want 10s", cfg.CommandTimeoutSetPower)
	}
	if cfg.CommandTimeoutSetChannel != 30*time.Second {
		t.Errorf("CommandTimeoutSetChannel = %v, want 30s", cfg.CommandTimeoutSetChannel)
	}

	// CB-TIMING §6.1
	if cfg.EventBufferSize != 50 {
		t.Errorf("EventBufferSize = %d, want 50", cfg.EventBufferSize)
	}
	if cfg.EventBufferRetention != 1*time.Hour {
		t.Errorf("EventBufferRetention = %v, want 1h", cfg.EventBufferRetention)
	}
}
