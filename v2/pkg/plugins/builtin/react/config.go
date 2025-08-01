package react

import (
	"fmt"
)

// Config holds plugin configuration
type Config struct {
	TypesOutput     string
	HooksOutput     string
	IncludeComments bool
	SourcePaths     []string
}

// ParseFrom parses configuration from map
func (c *Config) ParseFrom(config map[string]interface{}) error {
	if v, ok := config["typesOutput"].(string); ok {
		c.TypesOutput = v
	}

	if v, ok := config["hooksOutput"].(string); ok {
		c.HooksOutput = v
	}

	if v, ok := config["includeComments"].(bool); ok {
		c.IncludeComments = v
	}

	if v, ok := config["sourcePaths"].([]interface{}); ok {
		for _, path := range v {
			if s, ok := path.(string); ok {
				c.SourcePaths = append(c.SourcePaths, s)
			}
		}
	}

	return c.Validate()
}

// Validate configuration
func (c *Config) Validate() error {
	if c.TypesOutput == "" {
		return fmt.Errorf("typesOutput is required")
	}

	if c.HooksOutput == "" {
		return fmt.Errorf("hooksOutput is required")
	}

	return nil
}