package options

import (
	"encoding/json"
	"fmt"
	"time"
)

// PluginSystemConfig configures the plugin system
type PluginSystemConfig struct {
	// Enable/disable entire plugin system
	Enabled bool `json:"enabled"`
	
	// Plugin discovery paths (optional)
	DiscoveryPaths []string `json:"discoveryPaths,omitempty"`
	
	// Load timeout for plugins
	LoadTimeout time.Duration `json:"loadTimeout,omitempty"`
	
	// Individual plugin configurations
	Plugins map[string]*PluginConfig `json:"plugins,omitempty"`
}

// PluginConfig configures an individual plugin
type PluginConfig struct {
	// Enable/disable this specific plugin
	Enabled bool `json:"enabled"`
	
	// Plugin version constraint (optional)
	Version string `json:"version,omitempty"`
	
	// Plugin-specific configuration
	Config map[string]interface{} `json:"config,omitempty"`
	
	// Path for external plugins (optional)
	Path string `json:"path,omitempty"`
}

// Validate plugin configuration
func (p *PluginSystemConfig) Validate() error {
	if p.LoadTimeout < 0 {
		return fmt.Errorf("loadTimeout cannot be negative")
	}
	
	// Set defaults
	if p.LoadTimeout == 0 {
		p.LoadTimeout = 30 * time.Second
	}
	
	// Validate individual plugins
	for name, config := range p.Plugins {
		if name == "" {
			return fmt.Errorf("plugin name cannot be empty")
		}
		if config == nil {
			return fmt.Errorf("plugin %s has nil configuration", name)
		}
	}
	
	return nil
}

// SetDefaults applies default values
func (p *PluginSystemConfig) SetDefaults() {
	if p.LoadTimeout == 0 {
		p.LoadTimeout = 30 * time.Second
	}
	
	if p.DiscoveryPaths == nil {
		p.DiscoveryPaths = []string{
			"./.wails/plugins",
			"~/.wails/plugins",
		}
	}
}

// pluginSystemConfigJSON is a helper struct for JSON marshaling
type pluginSystemConfigJSON struct {
	Enabled        bool                     `json:"enabled"`
	DiscoveryPaths []string                 `json:"discoveryPaths,omitempty"`
	LoadTimeout    string                   `json:"loadTimeout,omitempty"`
	Plugins        map[string]*PluginConfig `json:"plugins,omitempty"`
}

// MarshalJSON custom marshaling for PluginSystemConfig
func (p *PluginSystemConfig) MarshalJSON() ([]byte, error) {
	aux := pluginSystemConfigJSON{
		Enabled:        p.Enabled,
		DiscoveryPaths: p.DiscoveryPaths,
		Plugins:        p.Plugins,
	}
	
	if p.LoadTimeout > 0 {
		aux.LoadTimeout = p.LoadTimeout.String()
	}
	
	return json.Marshal(aux)
}

// UnmarshalJSON custom unmarshaling for PluginSystemConfig
func (p *PluginSystemConfig) UnmarshalJSON(data []byte) error {
	var aux pluginSystemConfigJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	
	p.Enabled = aux.Enabled
	p.DiscoveryPaths = aux.DiscoveryPaths
	p.Plugins = aux.Plugins
	
	if aux.LoadTimeout != "" {
		duration, err := time.ParseDuration(aux.LoadTimeout)
		if err != nil {
			return fmt.Errorf("invalid loadTimeout format: %w", err)
		}
		p.LoadTimeout = duration
	}
	
	return nil
}