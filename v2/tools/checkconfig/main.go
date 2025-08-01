package main

import (
	"encoding/json"
	"fmt"
	"os"
	"io"

	"github.com/wailsapp/wails/v2/pkg/options"
)

func main() {
	// Read from stdin
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Printf("Error reading stdin: %v\n", err)
		os.Exit(1)
	}

	var app options.App
	err = json.Unmarshal(data, &app)
	if err != nil {
		fmt.Printf("Error unmarshaling config: %v\n", err)
		os.Exit(1)
	}

	// Validate the configuration
	err = app.Validate()
	if err != nil {
		fmt.Printf("Configuration validation failed: %v\n", err)
		os.Exit(1)
	}

	// Apply defaults
	options.MergeDefaults(&app)

	fmt.Println("Configuration is valid")
}