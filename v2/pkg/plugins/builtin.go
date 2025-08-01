package plugins

import (
	"context"
	"fmt"
)

// builtinRegistry is the global registry for built-in plugins
var builtinRegistry = NewRegistry()

// RegisterBuiltinPlugin registers a built-in plugin factory with the global registry.
// This should be called from init() functions in plugin packages.
func RegisterBuiltinPlugin(name string, factory PluginFactoryFunc) {
	if err := builtinRegistry.RegisterBuiltin(name, factory); err != nil {
		// Panic in init is acceptable for programming errors
		panic(fmt.Sprintf("failed to register built-in plugin %s: %v", name, err))
	}
}

// GetBuiltinRegistry returns the global built-in plugin registry.
func GetBuiltinRegistry() *Registry {
	return builtinRegistry
}

// Sample built-in plugins for basic framework support

// BasicReactPlugin provides basic React framework support.
type BasicReactPlugin struct {
	BasePlugin
}

// NewBasicReactPlugin creates a new React plugin instance.
func NewBasicReactPlugin() Plugin {
	return &BasicReactPlugin{
		BasePlugin: BasePlugin{
			name:        "basic-react",
			version:     "1.0.0",
			description: "Basic React framework support for Wails",
			author:      "Wails Team",
		},
	}
}

// Initialize initializes the React plugin.
func (p *BasicReactPlugin) Initialize(ctx context.Context, config *PluginConfig) error {
	if err := p.BasePlugin.Initialize(ctx, config); err != nil {
		return err
	}
	
	p.logger.Info("React plugin initialized", "projectRoot", config.ProjectRoot)
	return nil
}

// BasicVuePlugin provides basic Vue framework support.
type BasicVuePlugin struct {
	BasePlugin
}

// NewBasicVuePlugin creates a new Vue plugin instance.
func NewBasicVuePlugin() Plugin {
	return &BasicVuePlugin{
		BasePlugin: BasePlugin{
			name:        "basic-vue",
			version:     "1.0.0",
			description: "Basic Vue framework support for Wails",
			author:      "Wails Team",
		},
	}
}

// Initialize initializes the Vue plugin.
func (p *BasicVuePlugin) Initialize(ctx context.Context, config *PluginConfig) error {
	if err := p.BasePlugin.Initialize(ctx, config); err != nil {
		return err
	}
	
	p.logger.Info("Vue plugin initialized", "projectRoot", config.ProjectRoot)
	return nil
}

// DevToolsPlugin provides development tools integration.
type DevToolsPlugin struct {
	BasePlugin
	enabled bool
}

// NewDevToolsPlugin creates a new DevTools plugin instance.
func NewDevToolsPlugin() Plugin {
	return &DevToolsPlugin{
		BasePlugin: BasePlugin{
			name:        "dev-tools",
			version:     "1.0.0",
			description: "Development tools integration for Wails",
			author:      "Wails Team",
		},
	}
}

// Initialize initializes the DevTools plugin.
func (p *DevToolsPlugin) Initialize(ctx context.Context, config *PluginConfig) error {
	if err := p.BasePlugin.Initialize(ctx, config); err != nil {
		return err
	}
	
	// Check if dev tools should be enabled based on config
	if enabled, ok := config.Config["enabled"].(bool); ok {
		p.enabled = enabled
	} else {
		p.enabled = true // Default to enabled
	}
	
	p.logger.Info("DevTools plugin initialized", "enabled", p.enabled)
	return nil
}

// BasePlugin provides a base implementation for plugins.
// Other plugins can embed this to get default implementations.
type BasePlugin struct {
	name        string
	version     string
	description string
	author      string
	logger      Logger
	config      *PluginConfig
	initialized bool
}

// Name returns the plugin name.
func (p *BasePlugin) Name() string {
	return p.name
}

// Version returns the plugin version.
func (p *BasePlugin) Version() string {
	return p.version
}

// Description returns the plugin description.
func (p *BasePlugin) Description() string {
	return p.description
}

// Author returns the plugin author.
func (p *BasePlugin) Author() string {
	return p.author
}

// Initialize initializes the base plugin.
func (p *BasePlugin) Initialize(ctx context.Context, config *PluginConfig) error {
	if p.initialized {
		return nil
	}
	
	p.config = config
	p.logger = config.Logger
	p.initialized = true
	
	return nil
}

// Shutdown shuts down the base plugin.
func (p *BasePlugin) Shutdown(ctx context.Context) error {
	if !p.initialized {
		return nil
	}
	
	p.initialized = false
	p.logger.Debug("Plugin shutdown", "name", p.name)
	return nil
}

// Health checks the health of the base plugin.
func (p *BasePlugin) Health() error {
	if !p.initialized {
		return fmt.Errorf("plugin %s is not initialized", p.name)
	}
	return nil
}

// Register built-in plugins
func init() {
	// Register framework plugins
	RegisterBuiltinPlugin("basic-react", NewBasicReactPlugin)
	RegisterBuiltinPlugin("basic-vue", NewBasicVuePlugin)
	
	// Register tool plugins
	RegisterBuiltinPlugin("dev-tools", NewDevToolsPlugin)
	
	// Additional built-in plugins can be registered here
	// or in their respective packages' init functions
}

// LoadBuiltinPlugins loads all registered built-in plugins into the manager.
func LoadBuiltinPlugins(manager *Manager) error {
	registry := GetBuiltinRegistry()
	
	for _, name := range registry.ListBuiltins() {
		plugin, err := registry.CreateBuiltin(name)
		if err != nil {
			manager.logger.Error("Failed to create built-in plugin", "name", name, "error", err)
			continue
		}
		
		if err := manager.RegisterPlugin(plugin); err != nil {
			manager.logger.Error("Failed to register built-in plugin", "name", name, "error", err)
			continue
		}
	}
	
	return nil
}