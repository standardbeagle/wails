package plugins

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"
)

// ManagerConfig contains configuration for the plugin manager.
type ManagerConfig struct {
	// MaxInitTimeout is the maximum time allowed for plugin initialization
	MaxInitTimeout time.Duration

	// MaxShutdownTimeout is the maximum time allowed for plugin shutdown
	MaxShutdownTimeout time.Duration

	// EnableParallelHooks enables parallel execution of certain hooks
	EnableParallelHooks bool

	// Logger for manager operations
	Logger Logger

	// Project configuration
	ProjectRoot string
	ProjectName string

	// Plugin directories for discovery
	PluginDirs []string
}

// Manager manages the lifecycle and execution of plugins.
type Manager struct {
	// plugins stores all registered plugins by name
	plugins   map[string]Plugin
	pluginsMu sync.RWMutex

	// state tracks the current state of each plugin
	state   map[string]PluginState
	stateMu sync.RWMutex

	// config holds the manager configuration
	config *ManagerConfig

	// logger for manager operations
	logger Logger

	// initialized tracks if the manager has been initialized
	initialized bool
	initMu      sync.Mutex

	// registry for plugin discovery
	registry *Registry

	// Channels for coordinating shutdown
	shutdownCh chan struct{}
	shutdownWg sync.WaitGroup
}

// NewManager creates a new plugin manager with the given configuration.
func NewManager(config *ManagerConfig) *Manager {
	if config == nil {
		config = &ManagerConfig{
			MaxInitTimeout:      30 * time.Second,
			MaxShutdownTimeout:  10 * time.Second,
			EnableParallelHooks: true,
		}
	}

	// Set defaults if not provided
	if config.MaxInitTimeout == 0 {
		config.MaxInitTimeout = 30 * time.Second
	}
	if config.MaxShutdownTimeout == 0 {
		config.MaxShutdownTimeout = 10 * time.Second
	}

	logger := config.Logger
	if logger == nil {
		logger = &DefaultLogger{}
	}

	return &Manager{
		plugins:    make(map[string]Plugin),
		state:      make(map[string]PluginState),
		config:     config,
		logger:     logger,
		registry:   NewRegistry(),
		shutdownCh: make(chan struct{}),
	}
}

// Initialize initializes the manager and all registered plugins.
func (m *Manager) Initialize(ctx context.Context) error {
	m.initMu.Lock()
	defer m.initMu.Unlock()

	if m.initialized {
		return nil
	}

	m.logger.Info("Initializing plugin manager")

	// Discover plugins from configured directories
	if err := m.discoverPlugins(); err != nil {
		return fmt.Errorf("plugin discovery failed: %w", err)
	}

	// Initialize all registered plugins
	for name, plugin := range m.plugins {
		if err := m.initializePlugin(ctx, name, plugin); err != nil {
			m.logger.Error("Failed to initialize plugin", "plugin", name, "error", err)
			// Continue with other plugins even if one fails
			continue
		}
	}

	m.initialized = true
	m.logger.Info("Plugin manager initialized successfully")
	return nil
}

// RegisterPlugin registers a plugin with the manager.
func (m *Manager) RegisterPlugin(plugin Plugin) error {
	if plugin == nil {
		return &PluginError{
			Plugin:  "unknown",
			Message: "register",
			Cause:   fmt.Errorf("plugin is nil"),
		}
	}

	name := plugin.Name()
	if name == "" {
		return &PluginError{
			Plugin:  "unknown",
			Message: "register",
			Cause:   fmt.Errorf("plugin name is empty"),
		}
	}

	m.pluginsMu.Lock()
	defer m.pluginsMu.Unlock()

	if existing, exists := m.plugins[name]; exists {
		return &PluginError{
			Plugin:  name,
			Message: "register",
			Cause:   fmt.Errorf("plugin already registered with version %s", existing.Version()),
		}
	}

	m.plugins[name] = plugin
	m.setPluginState(name, StateUninitialized)
	
	m.logger.Debug("Plugin registered", "name", name, "version", plugin.Version())
	return nil
}

// GetPlugin returns a plugin by name.
func (m *Manager) GetPlugin(name string) (Plugin, error) {
	m.pluginsMu.RLock()
	defer m.pluginsMu.RUnlock()

	plugin, exists := m.plugins[name]
	if !exists {
		return nil, &PluginNotFoundError{Name: name}
	}

	return plugin, nil
}

// GetPluginsByType returns all plugins that implement the specified interface type.
// The interfaceType parameter should be a pointer to the interface type,
// e.g., (*BuildHook)(nil) for BuildHook interface.
func (m *Manager) GetPluginsByType(interfaceType interface{}) []Plugin {
	m.pluginsMu.RLock()
	defer m.pluginsMu.RUnlock()

	targetType := reflect.TypeOf(interfaceType).Elem()
	var plugins []Plugin

	for _, plugin := range m.plugins {
		if reflect.TypeOf(plugin).Implements(targetType) {
			// Only include plugins that are in ready state
			if m.getPluginState(plugin.Name()) == StateReady {
				plugins = append(plugins, plugin)
			}
		}
	}

	return plugins
}

// ExecuteHook executes a hook on all plugins that implement the specified interface.
// For BuildHook plugins, hookName would be "PreBuild" or "PostBuild".
func (m *Manager) ExecuteHook(hookName string, ctx interface{}) error {
	// Determine the interface type based on hook name
	var interfaceType interface{}
	switch hookName {
	case "PreBuild", "PostBuild":
		interfaceType = (*BuildHook)(nil)
	case "PreDev", "PostDev":
		interfaceType = (*DevHook)(nil)
	case "OnFileChanged":
		interfaceType = (*FileWatcher)(nil)
	case "GenerateCode":
		interfaceType = (*CodeGenerator)(nil)
	case "ProcessAssets":
		interfaceType = (*AssetProcessor)(nil)
	default:
		return fmt.Errorf("unknown hook: %s", hookName)
	}

	plugins := m.GetPluginsByType(interfaceType)
	if len(plugins) == 0 {
		m.logger.Debug("No plugins found for hook", "hook", hookName)
		return nil
	}

	if m.config.EnableParallelHooks && canRunParallel(hookName) {
		return m.executeHookParallel(plugins, hookName, ctx)
	}
	return m.executeHookSequential(plugins, hookName, ctx)
}

// Shutdown gracefully shuts down the manager and all plugins.
func (m *Manager) Shutdown(ctx context.Context) error {
	m.initMu.Lock()
	defer m.initMu.Unlock()

	if !m.initialized {
		return nil
	}

	m.logger.Info("Shutting down plugin manager")
	
	// Signal shutdown to any background operations
	close(m.shutdownCh)
	
	// Create a timeout context for shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, m.config.MaxShutdownTimeout)
	defer cancel()

	// Shutdown all plugins in reverse order of initialization
	var shutdownErrors []error
	
	m.pluginsMu.RLock()
	pluginNames := make([]string, 0, len(m.plugins))
	for name := range m.plugins {
		pluginNames = append(pluginNames, name)
	}
	m.pluginsMu.RUnlock()

	// Shutdown in reverse order
	for i := len(pluginNames) - 1; i >= 0; i-- {
		name := pluginNames[i]
		if err := m.shutdownPlugin(shutdownCtx, name); err != nil {
			shutdownErrors = append(shutdownErrors, err)
		}
	}

	// Wait for any background operations to complete
	m.shutdownWg.Wait()

	m.initialized = false
	m.logger.Info("Plugin manager shutdown complete")

	if len(shutdownErrors) > 0 {
		return combineErrors(shutdownErrors)
	}
	return nil
}

// Health checks the health of all plugins.
func (m *Manager) Health() error {
	m.pluginsMu.RLock()
	defer m.pluginsMu.RUnlock()

	var unhealthyPlugins []string
	for name, plugin := range m.plugins {
		if m.getPluginState(name) != StateReady {
			continue
		}

		if err := plugin.Health(); err != nil {
			unhealthyPlugins = append(unhealthyPlugins, fmt.Sprintf("%s: %v", name, err))
		}
	}

	if len(unhealthyPlugins) > 0 {
		return fmt.Errorf("unhealthy plugins: %v", unhealthyPlugins)
	}
	return nil
}

// GetPluginState returns the current state of a plugin.
func (m *Manager) GetPluginState(name string) (PluginState, error) {
	_, exists := m.plugins[name]
	if !exists {
		return StateUninitialized, &PluginNotFoundError{Name: name}
	}
	return m.getPluginState(name), nil
}

// Internal helper methods

func (m *Manager) initializePlugin(ctx context.Context, name string, plugin Plugin) error {
	m.logger.Debug("Initializing plugin", "name", name)
	
	// Update state
	if !m.updatePluginState(name, StateUninitialized, StateInitializing) {
		return &PluginStateError{
			Plugin:      name,
			CurrentState: m.getPluginState(name),
			RequestedState: StateInitializing,
		}
	}

	// Create plugin config
	config := &PluginConfig{
		ProjectRoot: m.config.ProjectRoot,
		ProjectName: m.config.ProjectName,
		Logger:      m.logger,
		Config:      make(map[string]interface{}), // TODO: Load from wails.json
	}

	// Initialize with timeout
	initCtx, cancel := context.WithTimeout(ctx, m.config.MaxInitTimeout)
	defer cancel()

	if err := plugin.Initialize(initCtx, config); err != nil {
		m.updatePluginState(name, StateInitializing, StateError)
		return &PluginInitError{
			Plugin: name,
			Cause:  err,
		}
	}

	// Update to ready state
	if !m.updatePluginState(name, StateInitializing, StateReady) {
		return &PluginStateError{
			Plugin:      name,
			CurrentState: m.getPluginState(name),
			RequestedState: StateReady,
		}
	}

	m.logger.Info("Plugin initialized successfully", "name", name, "version", plugin.Version())
	return nil
}

func (m *Manager) shutdownPlugin(ctx context.Context, name string) error {
	plugin, err := m.GetPlugin(name)
	if err != nil {
		return err
	}

	currentState := m.getPluginState(name)
	if currentState != StateReady {
		m.logger.Debug("Skipping shutdown for plugin not in ready state", "name", name, "state", currentState)
		return nil
	}

	m.logger.Debug("Shutting down plugin", "name", name)
	
	// Update state
	if !m.updatePluginState(name, StateReady, StateShuttingDown) {
		return &PluginStateError{
			Plugin:      name,
			CurrentState: m.getPluginState(name),
			RequestedState: StateShuttingDown,
		}
	}

	// Shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, m.config.MaxShutdownTimeout)
	defer cancel()

	if err := plugin.Shutdown(shutdownCtx); err != nil {
		m.updatePluginState(name, StateShuttingDown, StateError)
		return fmt.Errorf("plugin %s shutdown failed: %w", name, err)
	}

	// Update to stopped state
	if !m.updatePluginState(name, StateShuttingDown, StateStopped) {
		return &PluginStateError{
			Plugin:      name,
			CurrentState: m.getPluginState(name),
			RequestedState: StateStopped,
		}
	}

	m.logger.Info("Plugin shutdown successfully", "name", name)
	return nil
}

func (m *Manager) executeHookSequential(plugins []Plugin, hookName string, ctx interface{}) error {
	var errors []error
	
	for _, plugin := range plugins {
		if err := m.executePluginHook(plugin, hookName, ctx); err != nil {
			errors = append(errors, err)
			// Continue with other plugins even if one fails
		}
	}

	if len(errors) > 0 {
		return combineErrors(errors)
	}
	return nil
}

func (m *Manager) executeHookParallel(plugins []Plugin, hookName string, ctx interface{}) error {
	var wg sync.WaitGroup
	errors := make(chan error, len(plugins))

	for _, plugin := range plugins {
		wg.Add(1)
		go func(p Plugin) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					err := &PluginPanicError{
						Plugin:     p.Name(),
						PanicValue: r,
						Stack:      "stack trace not captured", // TODO: Add stack trace
					}
					errors <- fmt.Errorf("plugin %s panicked in hook %s: %v", p.Name(), hookName, err)
				}
			}()

			if err := m.executePluginHook(p, hookName, ctx); err != nil {
				errors <- err
			}
		}(plugin)
	}

	wg.Wait()
	close(errors)

	// Collect errors
	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return combineErrors(errs)
	}
	return nil
}

func (m *Manager) executePluginHook(plugin Plugin, hookName string, ctx interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			m.logger.Error("Plugin panicked", "plugin", plugin.Name(), "hook", hookName, "panic", r)
			err = &PluginPanicError{
				Plugin:     plugin.Name(),
				PanicValue: r,
				Stack:      "stack trace not captured", // TODO: Add stack trace
			}
		}
	}()

	// Use reflection to call the appropriate method
	method := reflect.ValueOf(plugin).MethodByName(hookName)
	if !method.IsValid() {
		return fmt.Errorf("plugin %s does not implement hook %s", plugin.Name(), hookName)
	}

	// Call the method with the context
	args := []reflect.Value{reflect.ValueOf(ctx)}
	results := method.Call(args)

	// Check for error return
	if len(results) > 0 && !results[len(results)-1].IsNil() {
		if err, ok := results[len(results)-1].Interface().(error); ok {
			return &PluginHookError{
				Plugin: plugin.Name(),
				Hook:   hookName,
				Err:    err,
			}
		}
	}

	return nil
}

func (m *Manager) getPluginState(name string) PluginState {
	m.stateMu.RLock()
	defer m.stateMu.RUnlock()
	
	state, exists := m.state[name]
	if !exists {
		return StateUninitialized
	}
	return state
}

func (m *Manager) setPluginState(name string, state PluginState) {
	m.stateMu.Lock()
	defer m.stateMu.Unlock()
	m.state[name] = state
}

func (m *Manager) updatePluginState(name string, from, to PluginState) bool {
	m.stateMu.Lock()
	defer m.stateMu.Unlock()

	current := m.state[name]
	if current != from {
		return false
	}

	if !from.IsValidTransition(to) {
		return false
	}

	m.state[name] = to
	return true
}

// canRunParallel determines if a hook can be executed in parallel.
func canRunParallel(hookName string) bool {
	// Some hooks should always run sequentially
	sequentialHooks := map[string]bool{
		"PreBuild":  true, // Build hooks might have dependencies
		"PostBuild": true,
	}
	
	return !sequentialHooks[hookName]
}

// combineErrors combines multiple errors into a single error.
func combineErrors(errors []error) error {
	if len(errors) == 0 {
		return nil
	}
	if len(errors) == 1 {
		return errors[0]
	}
	
	messages := make([]string, len(errors))
	for i, err := range errors {
		messages[i] = err.Error()
	}
	return fmt.Errorf("multiple errors: %s", strings.Join(messages, "; "))
}

// DefaultLogger provides a basic logger implementation.
type DefaultLogger struct{}

func (l *DefaultLogger) Debug(msg string, args ...interface{}) {
	// No-op for now
}

func (l *DefaultLogger) Info(msg string, args ...interface{}) {
	// No-op for now
}

func (l *DefaultLogger) Warn(msg string, args ...interface{}) {
	// No-op for now
}

func (l *DefaultLogger) Error(msg string, args ...interface{}) {
	// No-op for now
}

func (l *DefaultLogger) Fatal(msg string, args ...interface{}) {
	// No-op for now
}

func (l *DefaultLogger) WithField(key string, value interface{}) Logger {
	return l
}

func (l *DefaultLogger) WithFields(fields map[string]interface{}) Logger {
	return l
}