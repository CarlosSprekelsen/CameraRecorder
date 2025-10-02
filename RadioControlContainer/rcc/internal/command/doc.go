// Package command implements CommandOrchestrator from Architecture §5.
//
// Requirements:
//   - Architecture §5: "Validate requests (ranges, permissions); resolve channel index → frequency via ConfigStore; call adapter methods; emit events to TelemetryHub; write AuditLogger."
package command
