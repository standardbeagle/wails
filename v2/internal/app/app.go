package app

import (
	"context"
	"fmt"
	"time"

	"github.com/wailsapp/wails/v2/internal/frontend"
	"github.com/wailsapp/wails/v2/internal/logger"
	"github.com/wailsapp/wails/v2/internal/menumanager"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/plugins"
)

// App defines a Wails application structure
type App struct {
	frontend frontend.Frontend
	logger   *logger.Logger
	options  *options.App

	menuManager *menumanager.Manager

	// Indicates if the app is in debug mode
	debug bool

	// Indicates if the devtools is enabled
	devtoolsEnabled bool

	// OnStartup/OnShutdown
	startupCallback  func(ctx context.Context)
	shutdownCallback func(ctx context.Context)
	ctx              context.Context

	// Plugin manager (optional, nil when disabled)
	pluginManager *plugins.Manager
}

// Shutdown the application
func (a *App) Shutdown() {
	// Shutdown plugins first
	if a.pluginManager != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := a.pluginManager.Shutdown(ctx); err != nil {
			if a.logger != nil {
				a.logger.Error("Plugin shutdown error: %v", err)
			}
		}
	}

	if a.frontend != nil {
		a.frontend.Quit()
	}
}

// SetApplicationMenu sets the application menu
func (a *App) SetApplicationMenu(menu *menu.Menu) {
	if a.frontend != nil {
		a.frontend.MenuSetApplicationMenu(menu)
	}
}

// InitializePlugins initializes the plugin system if configured.
// This should be called during app initialization.
func (a *App) InitializePlugins(options *options.App) error {
	// Skip if plugins are not configured or disabled
	if options.Plugins == nil || !options.Plugins.Enabled {
		return nil
	}

	// Create logger for plugins
	pluginLogger := a.logger
	if pluginLogger == nil {
		// Use a no-op logger if no logger is available
		pluginLogger = &logger.Logger{}
	}

	// Create plugin manager configuration
	config := &plugins.ManagerConfig{
		MaxInitTimeout:      options.Plugins.LoadTimeout,
		MaxShutdownTimeout:  options.Plugins.ShutdownTimeout,
		EnableParallelHooks: options.Plugins.EnableParallelHooks,
		Logger:              &pluginLoggerAdapter{logger: pluginLogger},
		ProjectRoot:         "", // Will be set by caller
		ProjectName:         "", // Will be set by caller
		PluginDirs:          options.Plugins.DiscoveryPaths,
	}

	// Set defaults if not provided
	if config.MaxInitTimeout == 0 {
		config.MaxInitTimeout = 30 * time.Second
	}
	if config.MaxShutdownTimeout == 0 {
		config.MaxShutdownTimeout = 10 * time.Second
	}
	if len(config.PluginDirs) == 0 {
		config.PluginDirs = []string{"./plugins"}
	}

	// Create plugin manager
	a.pluginManager = plugins.NewManager(config)

	// Initialize plugins
	ctx, cancel := context.WithTimeout(context.Background(), config.MaxInitTimeout)
	defer cancel()

	if err := a.pluginManager.Initialize(ctx); err != nil {
		// Log error but don't fail app startup unless required
		if a.logger != nil {
			a.logger.Error("Plugin system initialization failed: %v", err)
		}

		if options.Plugins.Required {
			return fmt.Errorf("required plugin system failed: %w", err)
		}

		// Set plugin manager to nil if initialization failed and not required
		a.pluginManager = nil
	}

	return nil
}

// GetPluginManager returns the plugin manager if available.
func (a *App) GetPluginManager() *plugins.Manager {
	return a.pluginManager
}

// pluginLoggerAdapter adapts the Wails logger to the plugin logger interface.
type pluginLoggerAdapter struct {
	logger *logger.Logger
}

func (p *pluginLoggerAdapter) Debug(msg string, args ...interface{}) {
	if p.logger != nil {
		p.logger.Debug(fmt.Sprintf(msg, args...))
	}
}

func (p *pluginLoggerAdapter) Info(msg string, args ...interface{}) {
	if p.logger != nil {
		p.logger.Info(fmt.Sprintf(msg, args...))
	}
}

func (p *pluginLoggerAdapter) Warn(msg string, args ...interface{}) {
	if p.logger != nil {
		p.logger.Warning(fmt.Sprintf(msg, args...))
	}
}

func (p *pluginLoggerAdapter) Error(msg string, args ...interface{}) {
	if p.logger != nil {
		p.logger.Error(fmt.Sprintf(msg, args...))
	}
}

func (p *pluginLoggerAdapter) Fatal(msg string, args ...interface{}) {
	if p.logger != nil {
		p.logger.Fatal(fmt.Sprintf(msg, args...))
	}
}

func (p *pluginLoggerAdapter) WithField(key string, value interface{}) plugins.Logger {
	// For now, just return self as Wails logger doesn't support structured logging
	return p
}

func (p *pluginLoggerAdapter) WithFields(fields map[string]interface{}) plugins.Logger {
	// For now, just return self as Wails logger doesn't support structured logging
	return p
}

// ExecuteDevHook executes a development hook if plugins are available.
func (a *App) ExecuteDevHook(hookName string, devCtx *plugins.DevContext) error {
	if a.pluginManager == nil {
		return nil
	}

	if err := a.pluginManager.ExecuteHook(hookName, devCtx); err != nil {
		if a.logger != nil {
			a.logger.Warning("Dev hook '%s' errors: %v", hookName, err)
		}
		// Return error but don't fail the app
		return err
	}
	return nil
}

// ExecuteBuildHook executes a build hook if plugins are available.
func (a *App) ExecuteBuildHook(hookName string, buildCtx *plugins.BuildContext) error {
	if a.pluginManager == nil {
		return nil
	}

	if err := a.pluginManager.ExecuteHook(hookName, buildCtx); err != nil {
		if a.logger != nil {
			a.logger.Error("Build hook '%s' failed: %v", hookName, err)
		}
		// Build hooks should fail the build process
		return err
	}
	return nil
}
