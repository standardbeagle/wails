package options

import "time"

// PluginOptions contains plugin configuration for the application.
// This is an optional configuration that enables the plugin system.
type PluginOptions struct {
	// Enabled determines whether the plugin system is active
	Enabled bool `json:"enabled"`

	// Required determines whether plugin initialization failure should prevent app startup
	// If false, plugin errors are logged but don't stop the application
	Required bool `json:"required,omitempty"`

	// Plugins contains specific plugin configurations
	// Key is the plugin name, value is plugin-specific configuration
	Plugins map[string]*PluginEntry `json:"plugins,omitempty"`

	// DiscoveryPaths contains directories to search for plugins
	// Defaults to ["./plugins", "~/.wails/plugins"] if not specified
	DiscoveryPaths []string `json:"discoveryPaths,omitempty"`

	// LoadTimeout is the maximum time allowed for plugin initialization
	// Defaults to 30 seconds if not specified
	LoadTimeout time.Duration `json:"loadTimeout,omitempty"`

	// ShutdownTimeout is the maximum time allowed for plugin shutdown
	// Defaults to 10 seconds if not specified
	ShutdownTimeout time.Duration `json:"shutdownTimeout,omitempty"`

	// EnableParallelHooks enables parallel execution of certain hooks
	// Some hooks always run sequentially for safety
	EnableParallelHooks bool `json:"enableParallelHooks,omitempty"`
}

// PluginEntry represents configuration for a specific plugin.
type PluginEntry struct {
	// Enabled determines whether this specific plugin is active
	Enabled bool `json:"enabled"`

	// Version specifies the required plugin version (semver)
	Version string `json:"version,omitempty"`

	// Config contains plugin-specific configuration
	Config map[string]interface{} `json:"config,omitempty"`

	// Priority affects the order of plugin execution (higher = earlier)
	Priority int `json:"priority,omitempty"`

	// Path specifies a custom path to the plugin binary (for external plugins)
	Path string `json:"path,omitempty"`
}

// DefaultPluginOptions returns plugin options with sensible defaults.
func DefaultPluginOptions() *PluginOptions {
	return &PluginOptions{
		Enabled:             false, // Opt-in by default
		Required:            false, // Don't break apps if plugins fail
		Plugins:             make(map[string]*PluginEntry),
		DiscoveryPaths:      []string{"./plugins"},
		LoadTimeout:         30 * time.Second,
		ShutdownTimeout:     10 * time.Second,
		EnableParallelHooks: true,
	}
}

// IsPluginEnabled checks if a specific plugin is enabled.
func (p *PluginOptions) IsPluginEnabled(name string) bool {
	if !p.Enabled {
		return false
	}
	
	entry, exists := p.Plugins[name]
	if !exists {
		return false
	}
	
	return entry.Enabled
}

// GetPluginConfig returns configuration for a specific plugin.
func (p *PluginOptions) GetPluginConfig(name string) map[string]interface{} {
	if entry, exists := p.Plugins[name]; exists {
		return entry.Config
	}
	return nil
}