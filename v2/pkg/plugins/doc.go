/*
Package plugins provides the core plugin system interfaces and types for Wails.

This package defines the fundamental contracts that all Wails plugins must implement,
along with the supporting types, contexts, and error handling mechanisms needed
for a robust plugin architecture.

# Overview

The Wails plugin system is designed to be:
  - Extensible: Easy to add new plugin types and capabilities
  - Type-safe: Strong typing throughout the plugin lifecycle
  - Performant: Minimal overhead and efficient execution  
  - Maintainable: Clear interfaces and separation of concerns
  - Compatible: Forward and backward compatibility considerations

# Core Interfaces

The plugin system is built around several key interfaces:

Plugin: The base interface that all plugins must implement. Provides metadata
and lifecycle management.

    type Plugin interface {
        Name() string
        Version() string
        Description() string
        Author() string
        Initialize(ctx context.Context, config *PluginConfig) error
        Shutdown(ctx context.Context) error
        Health() error
    }

Specialized interfaces extend the base Plugin interface to provide specific
capabilities:

  - BuildHook: Provides hooks into the Wails build process
  - CodeGenerator: Generates code during the build process  
  - DevServerExtension: Extends the development server functionality
  - FileWatcher: Responds to file system changes during development
  - TemplateProvider: Provides project templates and initialization
  - AssetProcessor: Processes frontend assets during build

# Plugin Lifecycle

Plugins follow a well-defined lifecycle managed by the plugin system:

    1. Discovery: Plugins are discovered and registered
    2. Validation: Plugin metadata and interfaces are validated
    3. Initialization: Plugin Initialize() method is called with configuration
    4. Ready: Plugin is ready to handle requests and provide functionality
    5. Shutdown: Plugin Shutdown() method is called during application exit

Plugin states are tracked using the PluginState enumeration:

    StateUninitialized -> StateInitializing -> StateReady -> StateShuttingDown -> StateStopped

Error states can occur at any point and are represented by StateError.

# Context Types

The plugin system uses context types to pass information between the plugin
system and individual plugins:

  - BuildContext: Information about the current build process
  - GenerationContext: Data needed for code generation operations
  - FileChangeContext: Details about file system changes
  - AssetContext: Information about assets being processed
  - ProjectInitContext: Data for project template initialization
  - RuntimeContext: Information about the running application

All context types embed context.Context for cancellation and timeout support.

# Error Handling

The plugin system provides comprehensive error types for different failure scenarios:

  - PluginError: Generic plugin error with context
  - PluginNotFoundError: Plugin discovery failures
  - PluginInitError: Plugin initialization failures
  - PluginVersionError: Version compatibility issues
  - PluginDependencyError: Missing plugin dependencies
  - PluginConfigError: Configuration validation errors
  - PluginTimeoutError: Plugin operation timeouts
  - ValidationError: Multiple validation error aggregation

# Plugin Configuration

Plugins receive configuration through the PluginConfig struct, which includes:

  - User-provided configuration from wails.json
  - Project information (root path, name, etc.)
  - Environment variables and build settings
  - Logger instance for consistent logging
  - Wails configuration data

# Plugin Development

To create a new plugin, implement the Plugin interface and any additional
interfaces for specific capabilities:

    type MyPlugin struct {
        name string
        logger plugins.Logger
    }

    func (p *MyPlugin) Name() string { return p.name }
    func (p *MyPlugin) Version() string { return "1.0.0" }
    func (p *MyPlugin) Description() string { return "My custom plugin" }
    func (p *MyPlugin) Author() string { return "Plugin Author" }

    func (p *MyPlugin) Initialize(ctx context.Context, config *plugins.PluginConfig) error {
        p.logger = config.Logger
        p.logger.Info("Plugin initialized successfully")
        return nil
    }

    func (p *MyPlugin) Shutdown(ctx context.Context) error {
        p.logger.Info("Plugin shutting down")
        return nil
    }

    func (p *MyPlugin) Health() error {
        return nil // Plugin is healthy
    }

For specialized functionality, implement additional interfaces:

    // Implement BuildHook to participate in the build process
    func (p *MyPlugin) PreBuild(ctx *plugins.BuildContext) error {
        // Pre-build logic here
        return nil
    }

    func (p *MyPlugin) PostBuild(ctx *plugins.BuildContext) error {
        // Post-build logic here  
        return nil
    }

# Plugin Registration

Plugins are registered with the plugin manager during application startup:

    manager := plugins.NewManager()
    plugin := &MyPlugin{name: "my-plugin"}
    err := manager.Register(plugin)
    if err != nil {
        log.Fatal("Failed to register plugin:", err)
    }

# Thread Safety

All plugin interfaces and operations should be implemented in a thread-safe manner.
The plugin system may call plugin methods concurrently, especially during
development with file watching enabled.

# Best Practices

When developing plugins:

1. Always check for context cancellation in long-running operations
2. Use the provided logger for consistent logging output
3. Validate configuration early in the Initialize method
4. Clean up resources properly in the Shutdown method
5. Return meaningful errors with appropriate context
6. Document plugin capabilities and configuration options
7. Follow semantic versioning for plugin releases
8. Test plugins with various project configurations

# Compatibility

The plugin system is designed for forward compatibility. New interfaces may be
added in future versions, but existing interfaces will maintain backward
compatibility. Plugins should gracefully handle unknown configuration options
and provide sensible defaults.

*/
package plugins