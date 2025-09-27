package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Simple test to verify MediaMTX API functionality
func main() {
	// Test current MediaMTX API (should fail - API disabled)
	fmt.Println("Testing MediaMTX API on port 9997 (current)...")
	testAPI("http://localhost:9997/v3/config/get")

	// Test if we can access MediaMTX on port 8889 (WebRTC port)
	fmt.Println("\nTesting MediaMTX on port 8889 (WebRTC)...")
	testAPI("http://localhost:8889/v3/config/get")

	// Test if we can access MediaMTX on port 8888 (HLS port)
	fmt.Println("\nTesting MediaMTX on port 8888 (HLS)...")
	testAPI("http://localhost:8888/v3/config/get")
}

func testAPI(url string) {
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("  ❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("  Status: %s\n", resp.Status)
	fmt.Printf("  Content-Type: %s\n", resp.Header.Get("Content-Type"))

	// Try to parse as JSON
	var result interface{}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&result); err != nil {
		fmt.Printf("  ❌ Not JSON: %v\n", err)
	} else {
		fmt.Printf("  ✅ Valid JSON response\n")
	}
}
