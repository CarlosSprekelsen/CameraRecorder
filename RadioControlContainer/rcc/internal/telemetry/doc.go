// Package telemetry implements TelemetryHub from Architecture §5.
//
// Requirements:
//   - Architecture §5: "Fan-out events to all SSE clients; buffer last N events per client for reconnection (Last-Event-ID)."
package telemetry
