package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

func main() {
	// Set up Viper
	v := viper.New()
	v.SetConfigType("yaml")

	// Set environment variable handling
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetEnvPrefix("CAMERA_SERVICE")

	// Set a default value
	v.SetDefault("server.host", "default-host")

	// Set environment variable
	os.Setenv("CAMERA_SERVICE_SERVER_HOST", "   ")
	defer os.Unsetenv("CAMERA_SERVICE_SERVER_HOST")

	// Create a minimal config
	config := `
server:
  host: "default-host"
  port: 8002
`

	// Read from string
	v.ReadConfig(strings.NewReader(config))

	// Get the value
	host := v.GetString("server.host")
	fmt.Printf("Final host value: '%s'\n", host)
	fmt.Printf("Host value after TrimSpace: '%s'\n", strings.TrimSpace(host))
	fmt.Printf("Is empty after TrimSpace: %t\n", strings.TrimSpace(host) == "")
}
