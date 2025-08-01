package plugins

import (
	"fmt"
	"path/filepath"
	"sync"
)

// PluginFactoryFunc is a function that creates a new instance of a plugin.
type PluginFactoryFunc func() Plugin

// Registry manages plugin discovery and registration.
type Registry struct {
	// builtins stores factory functions for built-in plugins
	builtins   map[string]PluginFactoryFunc
	builtinsMu sync.RWMutex

	// external stores paths to external plugins
	external   map[string]string
	externalMu sync.RWMutex

	// metadata stores additional plugin metadata
	metadata   map[string]*PluginMetadata
	metadataMu sync.RWMutex
}

// NewRegistry creates a new plugin registry.
func NewRegistry() *Registry {
	return &Registry{
		builtins: make(map[string]PluginFactoryFunc),
		external: make(map[string]string),
		metadata: make(map[string]*PluginMetadata),
	}
}

// RegisterBuiltin registers a built-in plugin factory.
func (r *Registry) RegisterBuiltin(name string, factory PluginFactoryFunc) error {
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}
	if factory == nil {
		return fmt.Errorf("plugin factory cannot be nil")
	}

	r.builtinsMu.Lock()
	defer r.builtinsMu.Unlock()

	if _, exists := r.builtins[name]; exists {
		return fmt.Errorf("built-in plugin %s already registered", name)
	}

	r.builtins[name] = factory
	return nil
}

// RegisterExternal registers an external plugin by path.
func (r *Registry) RegisterExternal(name, path string) error {
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}
	if path == "" {
		return fmt.Errorf("plugin path cannot be empty")
	}

	// Validate path exists
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid plugin path: %w", err)
	}

	r.externalMu.Lock()
	defer r.externalMu.Unlock()

	if _, exists := r.external[name]; exists {
		return fmt.Errorf("external plugin %s already registered", name)
	}

	r.external[name] = absPath
	return nil
}

// RegisterMetadata registers metadata for a plugin.
func (r *Registry) RegisterMetadata(metadata *PluginMetadata) error {
	if metadata == nil {
		return fmt.Errorf("metadata cannot be nil")
	}
	if metadata.Name == "" {
		return fmt.Errorf("plugin name in metadata cannot be empty")
	}

	r.metadataMu.Lock()
	defer r.metadataMu.Unlock()

	r.metadata[metadata.Name] = metadata
	return nil
}

// GetBuiltin returns a built-in plugin factory by name.
func (r *Registry) GetBuiltin(name string) (PluginFactoryFunc, bool) {
	r.builtinsMu.RLock()
	defer r.builtinsMu.RUnlock()

	factory, exists := r.builtins[name]
	return factory, exists
}

// GetExternal returns an external plugin path by name.
func (r *Registry) GetExternal(name string) (string, bool) {
	r.externalMu.RLock()
	defer r.externalMu.RUnlock()

	path, exists := r.external[name]
	return path, exists
}

// GetMetadata returns metadata for a plugin by name.
func (r *Registry) GetMetadata(name string) (*PluginMetadata, bool) {
	r.metadataMu.RLock()
	defer r.metadataMu.RUnlock()

	metadata, exists := r.metadata[name]
	return metadata, exists
}

// ListBuiltins returns a list of all registered built-in plugin names.
func (r *Registry) ListBuiltins() []string {
	r.builtinsMu.RLock()
	defer r.builtinsMu.RUnlock()

	names := make([]string, 0, len(r.builtins))
	for name := range r.builtins {
		names = append(names, name)
	}
	return names
}

// ListExternal returns a list of all registered external plugin names.
func (r *Registry) ListExternal() []string {
	r.externalMu.RLock()
	defer r.externalMu.RUnlock()

	names := make([]string, 0, len(r.external))
	for name := range r.external {
		names = append(names, name)
	}
	return names
}

// CreateBuiltin creates a new instance of a built-in plugin.
func (r *Registry) CreateBuiltin(name string) (Plugin, error) {
	factory, exists := r.GetBuiltin(name)
	if !exists {
		return nil, &PluginNotFoundError{Name: name}
	}

	plugin := factory()
	if plugin == nil {
		return nil, fmt.Errorf("plugin factory for %s returned nil", name)
	}

	// Validate plugin metadata
	if plugin.Name() != name {
		return nil, fmt.Errorf("plugin name mismatch: expected %s, got %s", name, plugin.Name())
	}

	return plugin, nil
}

// LoadExternal loads an external plugin from the registered path.
// Note: This is a placeholder for future plugin loading functionality.
// Actual implementation would depend on the plugin format (shared library, separate process, etc.)
func (r *Registry) LoadExternal(name string) (Plugin, error) {
	path, exists := r.GetExternal(name)
	if !exists {
		return nil, &PluginNotFoundError{Name: name}
	}

	// TODO: Implement actual plugin loading based on plugin format
	// For now, return an error indicating this is not yet implemented
	return nil, fmt.Errorf("external plugin loading not yet implemented for %s at %s", name, path)
}

// Clear removes all registered plugins from the registry.
// This is mainly useful for testing.
func (r *Registry) Clear() {
	r.builtinsMu.Lock()
	r.builtins = make(map[string]PluginFactoryFunc)
	r.builtinsMu.Unlock()

	r.externalMu.Lock()
	r.external = make(map[string]string)
	r.externalMu.Unlock()

	r.metadataMu.Lock()
	r.metadata = make(map[string]*PluginMetadata)
	r.metadataMu.Unlock()
}

// RegistryStats provides statistics about the registry.
type RegistryStats struct {
	BuiltinCount  int
	ExternalCount int
	MetadataCount int
}

// Stats returns statistics about the registry.
func (r *Registry) Stats() RegistryStats {
	r.builtinsMu.RLock()
	builtinCount := len(r.builtins)
	r.builtinsMu.RUnlock()

	r.externalMu.RLock()
	externalCount := len(r.external)
	r.externalMu.RUnlock()

	r.metadataMu.RLock()
	metadataCount := len(r.metadata)
	r.metadataMu.RUnlock()

	return RegistryStats{
		BuiltinCount:  builtinCount,
		ExternalCount: externalCount,
		MetadataCount: metadataCount,
	}
}

// PluginSource represents where a plugin was loaded from.
type PluginSource int

const (
	// SourceBuiltin indicates the plugin is built into the application
	SourceBuiltin PluginSource = iota
	// SourceExternal indicates the plugin was loaded from an external source
	SourceExternal
	// SourceUnknown indicates the plugin source is unknown
	SourceUnknown
)

// GetPluginSource returns the source of a plugin by name.
func (r *Registry) GetPluginSource(name string) PluginSource {
	if _, exists := r.GetBuiltin(name); exists {
		return SourceBuiltin
	}
	if _, exists := r.GetExternal(name); exists {
		return SourceExternal
	}
	return SourceUnknown
}