/*
JWT Token Generator for MediaMTX Camera Service

This utility generates JWT tokens using the same secret key and algorithm
as the server, ensuring compatibility for testing and development.

Usage:
  go run main.go --role admin --expiry-hours 72
  go run main.go --role viewer --expiry-hours 24 --secret-key "custom-secret"
*/

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
)

var (
	role         = flag.String("role", "admin", "User role (viewer, operator, admin)")
	expiryHours  = flag.Int("expiry-hours", 48, "Token expiry in hours")
	secretKey    = flag.String("secret-key", "edge-device-secret-key-change-in-production", "JWT secret key")
	userID       = flag.String("user-id", "", "User ID (defaults to test_<role>)")
	outputFormat = flag.String("format", "token", "Output format: token, json")
)

func main() {
	flag.Parse()

	// Validate role
	if !security.ValidRoles[*role] {
		fmt.Fprintf(os.Stderr, "Error: Invalid role '%s'. Valid roles: viewer, operator, admin\n", *role)
		os.Exit(1)
	}

	// Validate expiry hours
	if *expiryHours <= 0 {
		fmt.Fprintf(os.Stderr, "Error: Expiry hours must be positive\n")
		os.Exit(1)
	}

	// Set default user ID if not provided
	if *userID == "" {
		*userID = "test_" + *role
	}

	// Create logger
	logger := logging.GetLogger("jwt-generator")

	// Create JWT handler
	jwtHandler, err := security.NewJWTHandler(*secretKey, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create JWT handler: %v\n", err)
		os.Exit(1)
	}

	// Generate token
	token, err := jwtHandler.GenerateToken(*userID, *role, *expiryHours)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to generate token: %v\n", err)
		os.Exit(1)
	}

	// Output result
	switch *outputFormat {
	case "json":
		expiresAt := time.Now().Add(time.Duration(*expiryHours) * time.Hour)
		output := fmt.Sprintf(`{
  "token": "%s",
  "user_id": "%s",
  "role": "%s",
  "expires_in_hours": %d,
  "expires_at": "%s",
  "algorithm": "HS256"
}`, token, *userID, *role, *expiryHours, expiresAt.Format(time.RFC3339))
		fmt.Println(output)
	case "token":
		fmt.Println(token)
	default:
		fmt.Fprintf(os.Stderr, "Error: Invalid output format '%s'. Valid formats: token, json\n", *outputFormat)
		os.Exit(1)
	}
}
