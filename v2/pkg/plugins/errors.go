package plugins

import (
	"fmt"
	"strings"
)

// PluginError represents a generic plugin error with context.
type PluginError struct {
	Plugin  string // Name of the plugin that caused the error
	Message string // Error message
	Cause   error  // Underlying error cause
	Code    string // Error code for programmatic handling
}

func (e *PluginError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("plugin '%s': %s: %v", e.Plugin, e.Message, e.Cause)
	}
	return fmt.Sprintf("plugin '%s': %s", e.Plugin, e.Message)
}

func (e *PluginError) Unwrap() error {
	return e.Cause
}

// NewPluginError creates a new plugin error.
func NewPluginError(plugin, message string, cause error) *PluginError {
	return &PluginError{
		Plugin:  plugin,
		Message: message,
		Cause:   cause,
	}
}

// NewPluginErrorWithCode creates a new plugin error with an error code.
func NewPluginErrorWithCode(plugin, message, code string, cause error) *PluginError {
	return &PluginError{
		Plugin:  plugin,
		Message: message,
		Cause:   cause,
		Code:    code,
	}
}

// PluginNotFoundError indicates a plugin could not be found.
type PluginNotFoundError struct {
	Name string // Name of the plugin that was not found
}

func (e *PluginNotFoundError) Error() string {
	return fmt.Sprintf("plugin not found: %s", e.Name)
}

// PluginInitError indicates a plugin failed to initialize.
type PluginInitError struct {
	Plugin string // Name of the plugin that failed to initialize
	Cause  error  // Underlying error cause
}

func (e *PluginInitError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("plugin '%s' failed to initialize: %v", e.Plugin, e.Cause)
	}
	return fmt.Sprintf("plugin '%s' failed to initialize", e.Plugin)
}

func (e *PluginInitError) Unwrap() error {
	return e.Cause
}

// PluginShutdownError indicates a plugin failed to shutdown cleanly.
type PluginShutdownError struct {
	Plugin string // Name of the plugin that failed to shutdown
	Cause  error  // Underlying error cause
}

func (e *PluginShutdownError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("plugin '%s' failed to shutdown: %v", e.Plugin, e.Cause)
	}
	return fmt.Sprintf("plugin '%s' failed to shutdown", e.Plugin)
}

func (e *PluginShutdownError) Unwrap() error {
	return e.Cause
}

// PluginVersionError indicates a plugin version compatibility issue.
type PluginVersionError struct {
	Plugin          string // Name of the plugin
	RequiredVersion string // Required version
	ActualVersion   string // Actual version found
}

func (e *PluginVersionError) Error() string {
	return fmt.Sprintf("plugin '%s' version mismatch: required %s, found %s", 
		e.Plugin, e.RequiredVersion, e.ActualVersion)
}

// PluginDependencyError indicates a plugin dependency issue.
type PluginDependencyError struct {
	Plugin       string   // Name of the plugin
	Dependencies []string // Missing dependencies
	Message      string   // Additional context
}

func (e *PluginDependencyError) Error() string {
	deps := strings.Join(e.Dependencies, ", ")
	if e.Message != "" {
		return fmt.Sprintf("plugin '%s' dependency error: missing [%s]: %s", 
			e.Plugin, deps, e.Message)
	}
	return fmt.Sprintf("plugin '%s' missing dependencies: [%s]", e.Plugin, deps)
}

// PluginConfigError indicates a plugin configuration error.
type PluginConfigError struct {
	Plugin string // Name of the plugin
	Field  string // Configuration field that caused the error
	Value  string // Invalid value
	Cause  error  // Underlying error cause
}

func (e *PluginConfigError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("plugin '%s' config error in field '%s' (value: %s): %v", 
			e.Plugin, e.Field, e.Value, e.Cause)
	}
	return fmt.Sprintf("plugin '%s' config error in field '%s' (value: %s)", 
		e.Plugin, e.Field, e.Value)
}

func (e *PluginConfigError) Unwrap() error {
	return e.Cause
}

// PluginTimeoutError indicates a plugin operation timed out.
type PluginTimeoutError struct {
	Plugin    string // Name of the plugin
	Operation string // Operation that timed out
	Timeout   string // Timeout duration
}

func (e *PluginTimeoutError) Error() string {
	return fmt.Sprintf("plugin '%s' operation '%s' timed out after %s", 
		e.Plugin, e.Operation, e.Timeout)
}

// PluginStateError indicates an invalid plugin state transition.
type PluginStateError struct {
	Plugin      string      // Name of the plugin
	CurrentState PluginState // Current state
	RequestedState PluginState // Requested state
	Message     string      // Additional context
}

func (e *PluginStateError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("plugin '%s' invalid state transition from %s to %s: %s", 
			e.Plugin, e.CurrentState, e.RequestedState, e.Message)
	}
	return fmt.Sprintf("plugin '%s' invalid state transition from %s to %s", 
		e.Plugin, e.CurrentState, e.RequestedState)
}

// ValidationError represents plugin validation errors.
type ValidationError struct {
	Plugin string   // Name of the plugin
	Errors []string // List of validation error messages
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("plugin '%s' validation failed: %s", 
		e.Plugin, strings.Join(e.Errors, "; "))
}

// AddError adds a validation error message.
func (e *ValidationError) AddError(message string) {
	e.Errors = append(e.Errors, message)
}

// HasErrors returns true if there are validation errors.
func (e *ValidationError) HasErrors() bool {
	return len(e.Errors) > 0
}

// NewValidationError creates a new validation error.
func NewValidationError(plugin string) *ValidationError {
	return &ValidationError{
		Plugin: plugin,
		Errors: make([]string, 0),
	}
}

// PluginPanicError wraps panics that occur in plugins.
type PluginPanicError struct {
	Plugin   string      // Name of the plugin
	PanicValue interface{} // Value passed to panic()
	Stack    string      // Stack trace
}

func (e *PluginPanicError) Error() string {
	return fmt.Sprintf("plugin '%s' panicked: %v", e.Plugin, e.PanicValue)
}

// IsPluginError checks if an error is a plugin-related error.
func IsPluginError(err error) bool {
	switch err.(type) {
	case *PluginError, *PluginNotFoundError, *PluginInitError, 
		 *PluginShutdownError, *PluginVersionError, *PluginDependencyError,
		 *PluginConfigError, *PluginTimeoutError, *PluginStateError,
		 *ValidationError, *PluginPanicError:
		return true
	default:
		return false
	}
}

// GetPluginName extracts the plugin name from a plugin error.
func GetPluginName(err error) string {
	switch e := err.(type) {
	case *PluginError:
		return e.Plugin
	case *PluginNotFoundError:
		return e.Name
	case *PluginInitError:
		return e.Plugin
	case *PluginShutdownError:
		return e.Plugin
	case *PluginVersionError:
		return e.Plugin
	case *PluginDependencyError:
		return e.Plugin
	case *PluginConfigError:
		return e.Plugin
	case *PluginTimeoutError:
		return e.Plugin
	case *PluginStateError:
		return e.Plugin
	case *ValidationError:
		return e.Plugin
	case *PluginPanicError:
		return e.Plugin
	default:
		return ""
	}
}