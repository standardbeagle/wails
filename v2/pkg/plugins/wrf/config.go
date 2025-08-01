package wrf

import (
	"fmt"
	"time"
)

// Config holds WRF plugin configuration
type Config struct {
	Go         GoConfig         `json:"go"`
	Generation GenerationConfig `json:"generation"`
	Development DevelopmentConfig `json:"development"`
}

// GoConfig configures Go source parsing
type GoConfig struct {
	SourcePaths []string `json:"sourcePaths"`
	WatchMode   bool     `json:"watchMode"`
	BuildTags   []string `json:"buildTags"`
}

// GenerationConfig configures code generation
type GenerationConfig struct {
	Hooks  HooksConfig  `json:"hooks"`
	Types  TypesConfig  `json:"types"`
	Events EventsConfig `json:"events"`
}

// HooksConfig configures hook generation
type HooksConfig struct {
	Output          string `json:"output"`
	Framework       string `json:"framework"` // "react-query", "swr", "basic"
	IncludeComments bool   `json:"includeComments"`
	CacheStrategies bool   `json:"cacheStrategies"`
}

// TypesConfig configures type generation
type TypesConfig struct {
	Output         string `json:"output"`
	Validation     string `json:"validation"` // "zod", "yup", "none"
	IncludeSchemas bool   `json:"includeSchemas"`
	NullableFields bool   `json:"nullableFields"`
}

// EventsConfig configures event type generation
type EventsConfig struct {
	Output string `json:"output"`
	Typed  bool   `json:"typed"`
}

// DevelopmentConfig configures development features
type DevelopmentConfig struct {
	HMR           bool `json:"hmr"`
	DevTools      bool `json:"devtools"`
	EventDebugging bool `json:"eventDebugging"`
}

// DefaultConfig returns default WRF configuration
func DefaultConfig() *Config {
	return &Config{
		Go: GoConfig{
			SourcePaths: []string{"./app/**/*.go"},
			WatchMode:   true,
			BuildTags:   []string{},
		},
		Generation: GenerationConfig{
			Hooks: HooksConfig{
				Output:          "frontend/src/hooks/generated",
				Framework:       "react-query",
				IncludeComments: true,
				CacheStrategies: true,
			},
			Types: TypesConfig{
				Output:         "frontend/src/types/generated.ts",
				Validation:     "zod",
				IncludeSchemas: true,
				NullableFields: true,
			},
			Events: EventsConfig{
				Output: "frontend/src/events/generated.ts",
				Typed:  true,
			},
		},
		Development: DevelopmentConfig{
			HMR:           true,
			DevTools:      true,
			EventDebugging: true,
		},
	}
}

// ParseFrom parses configuration from map
func (c *Config) ParseFrom(config map[string]interface{}) error {
	if goConfig, ok := config["go"].(map[string]interface{}); ok {
		if err := c.parseGoConfig(goConfig); err != nil {
			return fmt.Errorf("go config error: %w", err)
		}
	}

	if genConfig, ok := config["generation"].(map[string]interface{}); ok {
		if err := c.parseGenerationConfig(genConfig); err != nil {
			return fmt.Errorf("generation config error: %w", err)
		}
	}

	if devConfig, ok := config["development"].(map[string]interface{}); ok {
		if err := c.parseDevelopmentConfig(devConfig); err != nil {
			return fmt.Errorf("development config error: %w", err)
		}
	}

	return c.Validate()
}

func (c *Config) parseGoConfig(config map[string]interface{}) error {
	if paths, ok := config["sourcePaths"].([]interface{}); ok {
		c.Go.SourcePaths = make([]string, 0, len(paths))
		for _, path := range paths {
			if s, ok := path.(string); ok {
				c.Go.SourcePaths = append(c.Go.SourcePaths, s)
			}
		}
	}

	if watchMode, ok := config["watchMode"].(bool); ok {
		c.Go.WatchMode = watchMode
	}

	if tags, ok := config["buildTags"].([]interface{}); ok {
		c.Go.BuildTags = make([]string, 0, len(tags))
		for _, tag := range tags {
			if s, ok := tag.(string); ok {
				c.Go.BuildTags = append(c.Go.BuildTags, s)
			}
		}
	}

	return nil
}

func (c *Config) parseGenerationConfig(config map[string]interface{}) error {
	if hooks, ok := config["hooks"].(map[string]interface{}); ok {
		if output, ok := hooks["output"].(string); ok {
			c.Generation.Hooks.Output = output
		}
		if framework, ok := hooks["framework"].(string); ok {
			c.Generation.Hooks.Framework = framework
		}
		if includeComments, ok := hooks["includeComments"].(bool); ok {
			c.Generation.Hooks.IncludeComments = includeComments
		}
		if cacheStrategies, ok := hooks["cacheStrategies"].(bool); ok {
			c.Generation.Hooks.CacheStrategies = cacheStrategies
		}
	}

	if types, ok := config["types"].(map[string]interface{}); ok {
		if output, ok := types["output"].(string); ok {
			c.Generation.Types.Output = output
		}
		if validation, ok := types["validation"].(string); ok {
			c.Generation.Types.Validation = validation
		}
		if includeSchemas, ok := types["includeSchemas"].(bool); ok {
			c.Generation.Types.IncludeSchemas = includeSchemas
		}
		if nullableFields, ok := types["nullableFields"].(bool); ok {
			c.Generation.Types.NullableFields = nullableFields
		}
	}

	if events, ok := config["events"].(map[string]interface{}); ok {
		if output, ok := events["output"].(string); ok {
			c.Generation.Events.Output = output
		}
		if typed, ok := events["typed"].(bool); ok {
			c.Generation.Events.Typed = typed
		}
	}

	return nil
}

func (c *Config) parseDevelopmentConfig(config map[string]interface{}) error {
	if hmr, ok := config["hmr"].(bool); ok {
		c.Development.HMR = hmr
	}

	if devtools, ok := config["devtools"].(bool); ok {
		c.Development.DevTools = devtools
	}

	if eventDebugging, ok := config["eventDebugging"].(bool); ok {
		c.Development.EventDebugging = eventDebugging
	}

	return nil
}

// Validate configuration
func (c *Config) Validate() error {
	if len(c.Go.SourcePaths) == 0 {
		return fmt.Errorf("go.sourcePaths cannot be empty")
	}

	if c.Generation.Hooks.Output == "" {
		return fmt.Errorf("generation.hooks.output is required")
	}

	if c.Generation.Types.Output == "" {
		return fmt.Errorf("generation.types.output is required")
	}

	validFrameworks := []string{"react-query", "swr", "basic"}
	if !contains(validFrameworks, c.Generation.Hooks.Framework) {
		return fmt.Errorf("generation.hooks.framework must be one of: %v", validFrameworks)
	}

	validValidations := []string{"zod", "yup", "none"}
	if !contains(validValidations, c.Generation.Types.Validation) {
		return fmt.Errorf("generation.types.validation must be one of: %v", validValidations)
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}