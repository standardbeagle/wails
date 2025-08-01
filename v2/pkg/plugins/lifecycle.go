package plugins

import (
	"time"
)

// PluginConfig contains configuration data passed to plugins during initialization.
// This provides plugins with access to project information and user configuration.
type PluginConfig struct {
	// Plugin-specific configuration from wails.json or command line
	Config map[string]interface{} `json:"config"`

	// Project information
	ProjectRoot string `json:"project_root"` // Absolute path to project root
	ProjectName string `json:"project_name"` // Project name from wails.json
	
	// Wails configuration (interface{} to avoid circular dependency)
	WailsConfig interface{} `json:"wails_config"`
	
	// Environment information
	Environment string            `json:"environment"` // "development", "production", etc.
	Debug       bool              `json:"debug"`       // Debug mode enabled
	Variables   map[string]string `json:"variables"`   // Environment variables

	// Plugin system information
	PluginVersion string `json:"plugin_version"` // Plugin system version
	PluginDir     string `json:"plugin_dir"`     // Directory where plugins are stored

	// Logging interface
	Logger Logger `json:"-"` // Logger instance for this plugin

	// Startup time
	StartTime time.Time `json:"start_time"` // When the application started
}

// PluginState represents the current state of a plugin in its lifecycle.
type PluginState int

const (
	// StateUninitialized indicates the plugin has been created but not initialized
	StateUninitialized PluginState = iota
	
	// StateInitializing indicates the plugin is currently being initialized
	StateInitializing
	
	// StateReady indicates the plugin is initialized and ready to use
	StateReady
	
	// StateShuttingDown indicates the plugin is currently shutting down
	StateShuttingDown
	
	// StateStopped indicates the plugin has been shut down and is inactive
	StateStopped
	
	// StateError indicates the plugin is in an error state
	StateError
)

// String returns a human-readable representation of the plugin state.
func (s PluginState) String() string {
	switch s {
	case StateUninitialized:
		return "uninitialized"
	case StateInitializing:
		return "initializing"
	case StateReady:
		return "ready"
	case StateShuttingDown:
		return "shutting_down"
	case StateStopped:
		return "stopped"
	case StateError:
		return "error"
	default:
		return "unknown"
	}
}

// IsValidTransition checks if a state transition is valid.
func (s PluginState) IsValidTransition(to PluginState) bool {
	switch s {
	case StateUninitialized:
		return to == StateInitializing || to == StateError
	case StateInitializing:
		return to == StateReady || to == StateError
	case StateReady:
		return to == StateShuttingDown || to == StateError
	case StateShuttingDown:
		return to == StateStopped || to == StateError
	case StateStopped:
		return to == StateInitializing || to == StateError
	case StateError:
		return to == StateInitializing || to == StateStopped
	default:
		return false
	}
}

// PluginMetadata contains metadata about a plugin.
// This is used for plugin discovery and management.
type PluginMetadata struct {
	Name        string            `json:"name"`        // Plugin name
	Version     string            `json:"version"`     // Plugin version
	Description string            `json:"description"` // Plugin description
	Author      string            `json:"author"`      // Plugin author
	Homepage    string            `json:"homepage"`    // Plugin homepage URL
	Repository  string            `json:"repository"`  // Plugin repository URL
	License     string            `json:"license"`     // Plugin license
	Tags        []string          `json:"tags"`        // Plugin tags for categorization
	Keywords    []string          `json:"keywords"`    // Keywords for searching
	
	// Plugin capabilities
	Interfaces []string `json:"interfaces"` // Interfaces this plugin implements
	
	// Dependencies
	Dependencies     []string          `json:"dependencies"`      // Required plugins
	WailsVersion     string            `json:"wails_version"`     // Required Wails version
	GoVersion        string            `json:"go_version"`        // Required Go version
	PlatformSupport  []string          `json:"platform_support"`  // Supported platforms
	
	// Plugin configuration schema
	ConfigSchema map[string]interface{} `json:"config_schema"` // JSON schema for plugin config
	
	// Installation information
	InstallTime time.Time `json:"install_time"` // When plugin was installed
	UpdateTime  time.Time `json:"update_time"`  // When plugin was last updated
	Source      string    `json:"source"`       // Where plugin was installed from
	
	// Runtime information
	Enabled   bool                   `json:"enabled"`   // Whether plugin is enabled
	LoadTime  time.Duration          `json:"load_time"` // Time taken to load plugin
	State     PluginState            `json:"state"`     // Current plugin state
	LastError string                 `json:"last_error,omitempty"` // Last error message
	Metrics   map[string]interface{} `json:"metrics,omitempty"`    // Plugin metrics
}

// PluginRegistry defines methods for plugin registration and discovery.
type PluginRegistry interface {
	// Register a plugin instance
	Register(plugin Plugin) error
	
	// Unregister a plugin
	Unregister(name string) error
	
	// Get a plugin by name
	Get(name string) (Plugin, error)
	
	// List all registered plugins
	List() []Plugin
	
	// List plugins implementing a specific interface
	ListByInterface(interfaceName string) []Plugin
	
	// Get plugin metadata
	GetMetadata(name string) (*PluginMetadata, error)
}

// PluginLoader defines methods for loading plugins from various sources.
type PluginLoader interface {
	// Load plugin from file system
	LoadFromPath(path string) (Plugin, error)
	
	// Load plugin from Go plugin file
	LoadFromFile(filename string) (Plugin, error)
	
	// Load plugin from registry/repository
	LoadFromRegistry(name, version string) (Plugin, error)
	
	// Validate plugin before loading
	ValidatePlugin(plugin Plugin) error
}

// PluginManager defines the interface for managing the plugin lifecycle.
type PluginManager interface {
	PluginRegistry
	PluginLoader
	
	// Initialize all registered plugins
	InitializeAll() error
	
	// Initialize a specific plugin
	Initialize(name string, config *PluginConfig) error
	
	// Shutdown all plugins
	ShutdownAll() error
	
	// Shutdown a specific plugin
	Shutdown(name string) error
	
	// Enable/disable plugins
	Enable(name string) error
	Disable(name string) error
	
	// Health check for all plugins
	HealthCheck() map[string]error
	
	// Get plugin statistics
	GetStats() map[string]*PluginMetadata
}

// HookPriority defines the priority of plugin hooks.
// Lower numbers execute first.
type HookPriority int

const (
	PriorityHigh   HookPriority = 100
	PriorityNormal HookPriority = 500
	PriorityLow    HookPriority = 900
)

// HookRegistration represents a registered hook.
type HookRegistration struct {
	Plugin   string       // Plugin name
	Hook     string       // Hook name
	Priority HookPriority // Execution priority
	Handler  interface{}  // Hook handler function
	Enabled  bool         // Whether hook is enabled
}

// HookManager defines methods for managing plugin hooks.
type HookManager interface {
	// Register a hook
	RegisterHook(plugin, hook string, priority HookPriority, handler interface{}) error
	
	// Unregister a hook
	UnregisterHook(plugin, hook string) error
	
	// Execute all registered hooks for an event
	ExecuteHook(hook string, args ...interface{}) error
	
	// List registered hooks
	ListHooks(hook string) []*HookRegistration
	
	// Enable/disable specific hooks
	EnableHook(plugin, hook string) error
	DisableHook(plugin, hook string) error
}

// PluginFactory is a function that creates plugin instances.
type PluginFactory func() (Plugin, error)

// PluginDescriptor contains information needed to create a plugin instance.
type PluginDescriptor struct {
	Metadata *PluginMetadata // Plugin metadata
	Factory  PluginFactory   // Factory function to create plugin instances
	Config   *PluginConfig   // Default configuration
}