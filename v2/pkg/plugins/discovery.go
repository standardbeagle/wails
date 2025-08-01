package plugins

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// PluginManifest represents the metadata file for a plugin.
type PluginManifest struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description"`
	Author       string            `json:"author"`
	Type         string            `json:"type"` // "builtin", "external", "wasm", etc.
	EntryPoint   string            `json:"entryPoint,omitempty"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
	Config       json.RawMessage   `json:"config,omitempty"`
}

// DiscoveryOptions configures plugin discovery behavior.
type DiscoveryOptions struct {
	// Directories to search for plugins
	PluginDirs []string
	
	// Whether to include built-in plugins
	IncludeBuiltins bool
	
	// File patterns to match for plugin manifests
	ManifestPatterns []string
	
	// Whether to validate discovered plugins
	ValidatePlugins bool
}

// DefaultDiscoveryOptions returns default discovery options.
func DefaultDiscoveryOptions() *DiscoveryOptions {
	homeDir, _ := os.UserHomeDir()
	
	return &DiscoveryOptions{
		PluginDirs: []string{
			".wails/plugins",                          // Project-local plugins
			filepath.Join(homeDir, ".wails/plugins"),  // User plugins
		},
		IncludeBuiltins:  true,
		ManifestPatterns: []string{"plugin.json", "wails-plugin.json"},
		ValidatePlugins:  true,
	}
}

// discoverPlugins discovers and loads plugins based on the manager's configuration.
func (m *Manager) discoverPlugins() error {
	opts := DefaultDiscoveryOptions()
	
	// Add any additional directories from manager config
	if m.config.PluginDirs != nil {
		opts.PluginDirs = append(opts.PluginDirs, m.config.PluginDirs...)
	}
	
	// Load built-in plugins first
	if opts.IncludeBuiltins {
		if err := LoadBuiltinPlugins(m); err != nil {
			m.logger.Warn("Failed to load built-in plugins", "error", err)
		}
	}
	
	// Discover external plugins
	discovered, err := discoverExternalPlugins(opts)
	if err != nil {
		return fmt.Errorf("plugin discovery failed: %w", err)
	}
	
	// Register discovered plugins
	for _, manifest := range discovered {
		if err := m.registerDiscoveredPlugin(manifest); err != nil {
			m.logger.Error("Failed to register discovered plugin", 
				"name", manifest.Name, 
				"path", manifest.EntryPoint,
				"error", err)
			// Continue with other plugins
		}
	}
	
	m.logger.Info("Plugin discovery complete", 
		"builtins", len(m.registry.ListBuiltins()),
		"external", len(discovered))
	
	return nil
}

// discoverExternalPlugins searches for external plugins in configured directories.
func discoverExternalPlugins(opts *DiscoveryOptions) ([]*PluginManifest, error) {
	var manifests []*PluginManifest
	
	for _, dir := range opts.PluginDirs {
		// Skip if directory doesn't exist
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}
		
		// Search for plugin manifests
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip on error
			}
			
			// Check if file matches manifest patterns
			if !info.IsDir() && matchesPattern(info.Name(), opts.ManifestPatterns) {
				manifest, err := loadPluginManifest(path)
				if err != nil {
					return nil // Skip invalid manifests
				}
				
				// Set the directory as entry point if not specified
				if manifest.EntryPoint == "" {
					manifest.EntryPoint = filepath.Dir(path)
				}
				
				manifests = append(manifests, manifest)
			}
			
			return nil
		})
		
		if err != nil {
			return nil, fmt.Errorf("failed to walk directory %s: %w", dir, err)
		}
	}
	
	// Validate discovered plugins if requested
	if opts.ValidatePlugins {
		manifests = validateManifests(manifests)
	}
	
	return manifests, nil
}

// loadPluginManifest loads a plugin manifest from a JSON file.
func loadPluginManifest(path string) (*PluginManifest, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}
	
	var manifest PluginManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}
	
	// Validate required fields
	if manifest.Name == "" {
		return nil, fmt.Errorf("manifest missing required field: name")
	}
	if manifest.Version == "" {
		return nil, fmt.Errorf("manifest missing required field: version")
	}
	
	return &manifest, nil
}

// registerDiscoveredPlugin registers a plugin from a manifest.
func (m *Manager) registerDiscoveredPlugin(manifest *PluginManifest) error {
	// Convert manifest to plugin metadata
	metadata := &PluginMetadata{
		Name:        manifest.Name,
		Version:     manifest.Version,
		Description: manifest.Description,
		Author:      manifest.Author,
	}
	
	// Convert dependencies from map to slice if needed
	if manifest.Dependencies != nil {
		deps := make([]string, 0, len(manifest.Dependencies))
		for dep := range manifest.Dependencies {
			deps = append(deps, dep)
		}
		metadata.Dependencies = deps
	}
	
	// Register metadata
	if err := m.registry.RegisterMetadata(metadata); err != nil {
		return err
	}
	
	// Register based on type
	switch manifest.Type {
	case "external":
		// Register external plugin path
		if err := m.registry.RegisterExternal(manifest.Name, manifest.EntryPoint); err != nil {
			return err
		}
		
		// Try to load the plugin
		plugin, err := m.loadExternalPlugin(manifest)
		if err != nil {
			return fmt.Errorf("failed to load external plugin: %w", err)
		}
		
		return m.RegisterPlugin(plugin)
		
	case "wasm":
		// TODO: Implement WASM plugin loading
		return fmt.Errorf("WASM plugins not yet supported")
		
	default:
		return fmt.Errorf("unknown plugin type: %s", manifest.Type)
	}
}

// loadExternalPlugin loads an external plugin based on its manifest.
func (m *Manager) loadExternalPlugin(manifest *PluginManifest) (Plugin, error) {
	// TODO: Implement actual external plugin loading
	// This would involve:
	// 1. Loading shared library (.so/.dll/.dylib)
	// 2. Looking up plugin symbol
	// 3. Creating plugin instance
	// 4. Validating plugin interface
	
	return nil, fmt.Errorf("external plugin loading not yet implemented")
}

// matchesPattern checks if a filename matches any of the given patterns.
func matchesPattern(filename string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, filename); matched {
			return true
		}
	}
	return false
}

// validateManifests validates and filters plugin manifests.
func validateManifests(manifests []*PluginManifest) []*PluginManifest {
	seen := make(map[string]bool)
	valid := make([]*PluginManifest, 0, len(manifests))
	
	for _, manifest := range manifests {
		// Skip duplicates
		key := manifest.Name + "@" + manifest.Version
		if seen[key] {
			continue
		}
		seen[key] = true
		
		// Validate version format (basic semver check)
		if !isValidVersion(manifest.Version) {
			continue
		}
		
		valid = append(valid, manifest)
	}
	
	return valid
}

// isValidVersion performs basic semantic version validation.
func isValidVersion(version string) bool {
	// Basic check: should have format x.y.z
	parts := strings.Split(version, ".")
	return len(parts) >= 2 && len(parts) <= 3
}

// PluginDiscoveryResult represents the result of plugin discovery.
type PluginDiscoveryResult struct {
	BuiltinPlugins  []string
	ExternalPlugins []*PluginManifest
	Errors          []error
}

// DiscoverAvailablePlugins discovers all available plugins without loading them.
// This is useful for listing available plugins before initialization.
func DiscoverAvailablePlugins(opts *DiscoveryOptions) (*PluginDiscoveryResult, error) {
	if opts == nil {
		opts = DefaultDiscoveryOptions()
	}
	
	result := &PluginDiscoveryResult{
		BuiltinPlugins:  []string{},
		ExternalPlugins: []*PluginManifest{},
		Errors:          []error{},
	}
	
	// List built-in plugins
	if opts.IncludeBuiltins {
		registry := GetBuiltinRegistry()
		result.BuiltinPlugins = registry.ListBuiltins()
	}
	
	// Discover external plugins
	external, err := discoverExternalPlugins(opts)
	if err != nil {
		result.Errors = append(result.Errors, err)
	} else {
		result.ExternalPlugins = external
	}
	
	return result, nil
}