package plugins

import (
	"context"
	"time"
)

// BuildContext provides context for build-related plugin operations.
// Contains information about the current build process and project.
type BuildContext struct {
	context.Context

	// Project information
	ProjectRoot string // Absolute path to project root
	ProjectName string // Project name from wails.json
	OutputDir   string // Build output directory

	// Build configuration
	BuildMode    string            // "dev", "production", "debug"
	Platform     string            // Target platform (windows, darwin, linux)
	Architecture string            // Target architecture (amd64, arm64)
	Tags         []string          // Build tags
	LDFlags      []string          // Linker flags
	Environment  map[string]string // Environment variables

	// Wails configuration
	WailsConfig interface{} // Wails project configuration (avoid circular import)

	// Plugin-specific data
	PluginData map[string]interface{} // Data shared between plugins

	// Logging
	Logger Logger
}

// GenerationContext provides context for code generation operations.
// Contains information needed for generating code or other files.
type GenerationContext struct {
	context.Context

	// Project information
	ProjectRoot string // Absolute path to project root
	ProjectName string // Project name
	OutputDir   string // Where to write generated files

	// Generation parameters
	Language    string            // Target language for generation
	Framework   string            // Frontend framework (if applicable)
	Templates   []string          // Template paths to use
	Variables   map[string]string // Template variables
	Options     map[string]interface{} // Generator-specific options

	// Source information
	SourceFiles []string           // Source files to analyze
	Metadata    map[string]interface{} // Extracted metadata

	// Plugin data
	PluginData map[string]interface{} // Data from other plugins

	// Logging  
	Logger Logger
}

// FileChangeContext provides context for file change events.
// Contains information about what files changed and how.
type FileChangeContext struct {
	context.Context

	// Project information
	ProjectRoot string // Absolute path to project root

	// Change information
	FilePath   string    // Path of changed file (relative to project root)
	ChangeType string    // "created", "modified", "deleted", "renamed"
	OldPath    string    // Previous path (for rename events)
	Timestamp  time.Time // When the change occurred

	// File information
	IsDirectory bool   // Whether the changed path is a directory
	FileSize    int64  // Size of the file (0 for directories)
	Checksum    string // File checksum (for change detection)

	// Context data
	PluginData map[string]interface{} // Data from other plugins
	
	// Logging
	Logger Logger
}

// AssetContext provides context for asset processing operations.
// Contains information about assets being processed.
type AssetContext struct {
	context.Context

	// Project information
	ProjectRoot string // Absolute path to project root
	AssetsDir   string // Assets directory path
	OutputDir   string // Where to write processed assets

	// Asset information
	AssetPath   string            // Path to asset being processed
	AssetType   string            // Asset type (css, js, image, etc.)
	ContentType string            // MIME type
	Metadata    map[string]string // Asset metadata

	// Processing options
	Minify     bool              // Whether to minify output
	SourceMaps bool              // Whether to generate source maps
	Options    map[string]interface{} // Processor-specific options

	// Dependencies
	Dependencies []string // Asset dependencies

	// Plugin data
	PluginData map[string]interface{} // Data from other plugins

	// Logging
	Logger Logger
}

// ProjectInitContext provides context for project initialization.
// Contains information needed to initialize a new project from a template.
type ProjectInitContext struct {
	context.Context

	// Project information
	ProjectRoot string // Absolute path to new project directory
	ProjectName string // Name of the new project
	ProjectPath string // Relative path within workspace

	// Template information
	TemplateName string            // Name of template being used
	Variables    map[string]string // Template variables from user input
	Options      map[string]interface{} // Template-specific options

	// User information
	Author      string // Project author
	Description string // Project description
	License     string // Project license

	// Framework selection
	Frontend string // Frontend framework choice
	Backend  string // Backend language/framework

	// Plugin data
	PluginData map[string]interface{} // Data from other plugins

	// Logging
	Logger Logger
}

// RuntimeContext provides context for runtime plugin operations.
// Contains information about the running application.
type RuntimeContext struct {
	context.Context

	// Application information
	AppName    string // Application name
	AppVersion string // Application version
	BuildInfo  string // Build information

	// Runtime environment
	Platform     string // Runtime platform
	Architecture string // Runtime architecture
	WorkingDir   string // Current working directory

	// Application state
	Windows    []WindowInfo           // Open application windows
	Processes  []ProcessInfo          // Running processes
	Memory     MemoryInfo             // Memory usage information
	Config     map[string]interface{} // Application configuration

	// Plugin data
	PluginData map[string]interface{} // Data shared between plugins

	// Logging
	Logger Logger
}

// WindowInfo contains information about an application window.
type WindowInfo struct {
	ID       string // Window identifier
	Title    string // Window title
	Width    int    // Window width
	Height   int    // Window height
	X        int    // Window X position
	Y        int    // Window Y position
	Visible  bool   // Whether window is visible
	Focused  bool   // Whether window has focus
	Metadata map[string]interface{} // Additional window metadata
}

// ProcessInfo contains information about a running process.
type ProcessInfo struct {
	PID      int                    // Process ID
	Name     string                 // Process name
	Command  string                 // Command line
	Memory   int64                  // Memory usage in bytes
	CPU      float64                // CPU usage percentage
	Metadata map[string]interface{} // Additional process metadata
}

// MemoryInfo contains memory usage information.
type MemoryInfo struct {
	Total     int64 // Total memory in bytes
	Used      int64 // Used memory in bytes
	Available int64 // Available memory in bytes
	Swap      int64 // Swap usage in bytes
}

// Logger defines the interface for plugin logging.
// Allows plugins to log messages through the Wails logging system.
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Fatal(msg string, args ...interface{})
	
	// Structured logging
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
}