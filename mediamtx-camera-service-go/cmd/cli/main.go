/*
CLI Utility for MediaMTX Camera Service

Requirements Coverage:
- REQ-SEC-014: Key Management
- REQ-SEC-015: Production API Key Management

Test Categories: Unit/Integration
API Documentation Reference: docs/api/json_rpc_methods.md

Provides command-line interface for API key management operations.
Follows canonical configuration patterns and existing CLI patterns.
*/

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
)

const (
	appName    = "camera-service-cli"
	appVersion = "1.0.0"
)

var (
	configPath = flag.String("config", "/etc/camera-service/config.yaml", "Path to configuration file")
	verbose    = flag.Bool("verbose", false, "Enable verbose output")
	format     = flag.String("format", "table", "Output format (table, json)")
)

func main() {
	flag.Parse()

	// Setup logging
	logger := logging.GetLogger("cli")
	if *verbose {
		logger.SetLevel(logging.DebugLevel)
	}

	// Load configuration
	configManager := config.CreateConfigManager()
	if err := configManager.LoadConfig(*configPath); err != nil {
		logger.WithError(err).Fatal("Failed to load configuration")
	}

	cfg := configManager.GetConfig()
	if cfg == nil {
		logger.Fatal("Configuration is nil")
	}

	// Check if API key management is enabled
	if !cfg.APIKeyManagement.CLIEnabled {
		logger.Fatal("API key management CLI is disabled in configuration")
	}

	// Create API key manager
	keyManager, err := security.NewAPIKeyManager(&cfg.APIKeyManagement, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create API key manager")
	}

	// Parse command
	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	command := args[0]
	commandArgs := args[1:]

	// Execute command
	ctx := context.Background()
	if err := executeCommand(ctx, keyManager, command, commandArgs); err != nil {
		logger.WithError(err).Fatal("Command execution failed")
	}
}

func executeCommand(ctx context.Context, keyManager *security.APIKeyManager, command string, args []string) error {
	switch command {
	case "keys":
		return executeKeysCommand(ctx, keyManager, args)
	case "version":
		printVersion()
		return nil
	case "help":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func executeKeysCommand(ctx context.Context, keyManager *security.APIKeyManager, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("keys command requires a subcommand")
	}

	subcommand := args[0]
	subcommandArgs := args[1:]

	switch subcommand {
	case "generate":
		return executeKeysGenerate(ctx, keyManager, subcommandArgs)
	case "list":
		return executeKeysList(ctx, keyManager, subcommandArgs)
	case "revoke":
		return executeKeysRevoke(ctx, keyManager, subcommandArgs)
	case "rotate":
		return executeKeysRotate(ctx, keyManager, subcommandArgs)
	case "export":
		return executeKeysExport(ctx, keyManager, subcommandArgs)
	case "cleanup":
		return executeKeysCleanup(ctx, keyManager, subcommandArgs)
	case "stats":
		return executeKeysStats(ctx, keyManager, subcommandArgs)
	default:
		return fmt.Errorf("unknown keys subcommand: %s", subcommand)
	}
}

func executeKeysGenerate(ctx context.Context, keyManager *security.APIKeyManager, args []string) error {
	// Parse flags
	fs := flag.NewFlagSet("keys generate", flag.ExitOnError)
	role := fs.String("role", "", "Role for the API key (viewer, operator, admin)")
	expiry := fs.String("expiry", "90d", "Key expiry duration (e.g., 90d, 30d, 1y)")
	description := fs.String("description", "", "Description for the API key")
	force := fs.Bool("force", false, "Force generation even if max keys per role exceeded")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Validate required flags
	if *role == "" {
		return fmt.Errorf("role is required")
	}

	// Parse expiry duration
	duration, err := parseDuration(*expiry)
	if err != nil {
		return fmt.Errorf("invalid expiry duration: %w", err)
	}

	// Validate role
	validRole, err := security.ValidateRole(*role)
	if err != nil {
		return fmt.Errorf("invalid role: %w", err)
	}

	// Generate key
	apiKey, err := keyManager.GenerateKey(validRole, duration, *description)
	if err != nil {
		return fmt.Errorf("failed to generate API key: %w", err)
	}

	// Output result
	if *format == "json" {
		output, _ := json.MarshalIndent(apiKey, "", "  ")
		fmt.Println(string(output))
	} else {
		fmt.Printf("API Key Generated Successfully:\n")
		fmt.Printf("  ID:          %s\n", apiKey.ID)
		fmt.Printf("  Key:         %s\n", apiKey.Key)
		fmt.Printf("  Role:        %s\n", apiKey.Role)
		fmt.Printf("  Expires:     %s\n", apiKey.ExpiresAt.Format(time.RFC3339))
		fmt.Printf("  Description: %s\n", apiKey.Description)
		fmt.Printf("\n⚠️  IMPORTANT: Store this key securely. It will not be shown again.\n")
	}

	return nil
}

func executeKeysList(ctx context.Context, keyManager *security.APIKeyManager, args []string) error {
	// Parse flags
	fs := flag.NewFlagSet("keys list", flag.ExitOnError)
	role := fs.String("role", "", "Filter by role (viewer, operator, admin)")
	status := fs.String("status", "", "Filter by status (active, revoked, expired)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Validate role if provided
	var validRole security.Role
	if *role != "" {
		var err error
		validRole, err = security.ValidateRole(*role)
		if err != nil {
			return fmt.Errorf("invalid role: %w", err)
		}
	}

	// List keys
	keys, err := keyManager.ListKeys(validRole)
	if err != nil {
		return fmt.Errorf("failed to list API keys: %w", err)
	}

	// Filter by status if specified
	if *status != "" {
		var filteredKeys []*security.APIKey
		for _, key := range keys {
			if key.Status == *status {
				filteredKeys = append(filteredKeys, key)
			}
		}
		keys = filteredKeys
	}

	// Output result
	if *format == "json" {
		output, _ := json.MarshalIndent(keys, "", "  ")
		fmt.Println(string(output))
	} else {
		if len(keys) == 0 {
			fmt.Println("No API keys found")
			return nil
		}

		fmt.Printf("API Keys (%d found):\n\n", len(keys))
		fmt.Printf("%-36s %-10s %-10s %-20s %-30s\n", "ID", "Role", "Status", "Expires", "Description")
		fmt.Printf("%s\n", strings.Repeat("-", 106))

		for _, key := range keys {
			expiresStr := key.ExpiresAt.Format("2006-01-02 15:04:05")
			if key.ExpiresAt.IsZero() {
				expiresStr = "Never"
			}
			fmt.Printf("%-36s %-10s %-10s %-20s %-30s\n",
				key.ID, key.Role, key.Status, expiresStr, key.Description)
		}
	}

	return nil
}

func executeKeysRevoke(ctx context.Context, keyManager *security.APIKeyManager, args []string) error {
	// Parse flags
	fs := flag.NewFlagSet("keys revoke", flag.ExitOnError)
	keyID := fs.String("key-id", "", "ID of the key to revoke")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Validate required flags
	if *keyID == "" {
		return fmt.Errorf("key-id is required")
	}

	// Revoke key
	if err := keyManager.RevokeKey(*keyID); err != nil {
		return fmt.Errorf("failed to revoke API key: %w", err)
	}

	fmt.Printf("API key %s revoked successfully\n", *keyID)
	return nil
}

func executeKeysRotate(ctx context.Context, keyManager *security.APIKeyManager, args []string) error {
	// Parse flags
	fs := flag.NewFlagSet("keys rotate", flag.ExitOnError)
	role := fs.String("role", "", "Role to rotate keys for (viewer, operator, admin)")
	force := fs.Bool("force", false, "Force rotation even if no keys exist")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Validate required flags
	if *role == "" {
		return fmt.Errorf("role is required")
	}

	// Validate role
	validRole, err := security.ValidateRole(*role)
	if err != nil {
		return fmt.Errorf("invalid role: %w", err)
	}

	// Rotate keys
	if err := keyManager.RotateKeys(validRole, *force); err != nil {
		return fmt.Errorf("failed to rotate API keys: %w", err)
	}

	fmt.Printf("API keys for role %s rotated successfully\n", *role)
	return nil
}

func executeKeysExport(ctx context.Context, keyManager *security.APIKeyManager, args []string) error {
	// Parse flags
	fs := flag.NewFlagSet("keys export", flag.ExitOnError)
	keyID := fs.String("key-id", "", "ID of the key to export")
	output := fs.String("output", "", "Output file path (default: stdout)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Validate required flags
	if *keyID == "" {
		return fmt.Errorf("key-id is required")
	}

	// Get key details
	keys, err := keyManager.ListKeys("")
	if err != nil {
		return fmt.Errorf("failed to list API keys: %w", err)
	}

	var targetKey *security.APIKey
	for _, key := range keys {
		if key.ID == *keyID {
			targetKey = key
			break
		}
	}

	if targetKey == nil {
		return fmt.Errorf("key not found: %s", *keyID)
	}

	// Export key
	exportData := map[string]interface{}{
		"key":        targetKey,
		"exported_at": time.Now().Format(time.RFC3339),
		"exported_by": "camera-service-cli",
	}

	outputData, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal export data: %w", err)
	}

	// Write output
	if *output != "" {
		if err := os.WriteFile(*output, outputData, 0600); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Key exported to %s\n", *output)
	} else {
		fmt.Println(string(outputData))
	}

	return nil
}

func executeKeysCleanup(ctx context.Context, keyManager *security.APIKeyManager, args []string) error {
	// Parse flags
	fs := flag.NewFlagSet("keys cleanup", flag.ExitOnError)
	dryRun := fs.Bool("dry-run", false, "Show what would be cleaned up without making changes")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *dryRun {
		// List expired keys
		keys, err := keyManager.ListKeys("")
		if err != nil {
			return fmt.Errorf("failed to list API keys: %w", err)
		}

		var expiredKeys []*security.APIKey
		now := time.Now()
		for _, key := range keys {
			if now.After(key.ExpiresAt) && key.Status == "active" {
				expiredKeys = append(expiredKeys, key)
			}
		}

		if len(expiredKeys) == 0 {
			fmt.Println("No expired keys found")
			return nil
		}

		fmt.Printf("Expired keys that would be cleaned up (%d found):\n\n", len(expiredKeys))
		for _, key := range expiredKeys {
			fmt.Printf("  %s (%s) - expired %s\n", key.ID, key.Role, key.ExpiresAt.Format("2006-01-02 15:04:05"))
		}
	} else {
		// Cleanup expired keys
		if err := keyManager.CleanupExpiredKeys(); err != nil {
			return fmt.Errorf("failed to cleanup expired keys: %w", err)
		}
		fmt.Println("Expired keys cleaned up successfully")
	}

	return nil
}

func executeKeysStats(ctx context.Context, keyManager *security.APIKeyManager, args []string) error {
	// Get statistics
	stats := keyManager.GetStats()

	// Output result
	if *format == "json" {
		output, _ := json.MarshalIndent(stats, "", "  ")
		fmt.Println(string(output))
	} else {
		fmt.Println("API Key Statistics:")
		fmt.Printf("  Total Keys:   %d\n", stats["total_keys"])
		fmt.Printf("  Active Keys:  %d\n", stats["active_keys"])
		fmt.Printf("  Revoked Keys: %d\n", stats["revoked_keys"])
		fmt.Printf("  Expired Keys: %d\n", stats["expired_keys"])

		keysByRole := stats["keys_by_role"].(map[string]int)
		if len(keysByRole) > 0 {
			fmt.Println("\nKeys by Role:")
			for role, count := range keysByRole {
				fmt.Printf("  %s: %d\n", role, count)
			}
		}
	}

	return nil
}

func parseDuration(s string) (time.Duration, error) {
	// Parse duration string (e.g., "90d", "30d", "1y")
	if strings.HasSuffix(s, "d") {
		days := s[:len(s)-1]
		return time.ParseDuration(days + "h" + "24")
	} else if strings.HasSuffix(s, "y") {
		years := s[:len(s)-1]
		return time.ParseDuration(years + "h" + "8760")
	} else {
		return time.ParseDuration(s)
	}
}

func printUsage() {
	fmt.Printf(`%s - MediaMTX Camera Service CLI

Usage:
  %s [flags] <command> [command-flags]

Commands:
  keys <subcommand>    Manage API keys
  version              Show version information
  help                 Show this help message

Key Management Commands:
  keys generate        Generate a new API key
  keys list            List existing API keys
  keys revoke          Revoke an API key
  keys rotate          Rotate keys for a role
  keys export          Export a key for backup
  keys cleanup         Clean up expired keys
  keys stats           Show key statistics

Flags:
  -config string       Path to configuration file (default: /etc/camera-service/config.yaml)
  -verbose             Enable verbose output
  -format string       Output format: table or json (default: table)

Examples:
  %s keys generate --role admin --expiry 90d --description "Production admin key"
  %s keys list --role admin --status active
  %s keys revoke --key-id abc123
  %s keys rotate --role operator --force
  %s keys export --key-id abc123 --output backup.json
  %s keys cleanup --dry-run
  %s keys stats

`, appName, appName, appName, appName, appName, appName, appName, appName, appName)
}

func printVersion() {
	fmt.Printf("%s version %s\n", appName, appVersion)
}
