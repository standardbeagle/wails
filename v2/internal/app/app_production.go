//go:build production

package app

import (
	"context"

	"github.com/wailsapp/wails/v2/internal/binding"
	"github.com/wailsapp/wails/v2/internal/frontend/desktop"
	"github.com/wailsapp/wails/v2/internal/frontend/dispatcher"
	"github.com/wailsapp/wails/v2/internal/frontend/runtime"
	"github.com/wailsapp/wails/v2/internal/logger"
	"github.com/wailsapp/wails/v2/internal/menumanager"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/plugins"
)

func (a *App) Run() error {
	// Execute pre-build hooks for production
	if a.pluginManager != nil {
		buildCtx := &plugins.BuildContext{
			Context:      context.Background(),
			ProjectRoot:  "", // Will be set by build system
			ProjectName:  "", // Will be set by build system
			OutputDir:    "", // Will be set by build system
			BuildMode:    "production",
			Platform:     "", // Will be set by build system
			Architecture: "", // Will be set by build system
			Tags:         []string{},
			LDFlags:      []string{},
			Environment:  make(map[string]string),
			WailsConfig:  a.options,
			PluginData:   make(map[string]interface{}),
			Logger:       &pluginLoggerAdapter{logger: a.logger},
		}

		if err := a.ExecuteBuildHook("PreBuild", buildCtx); err != nil {
			a.logger.Error("Pre-build hook failed: %v", err)
			return err
		}

		// Store build context for post hooks
		a.ctx = context.WithValue(a.ctx, "buildCtx", buildCtx)
	}

	err := a.frontend.Run(a.ctx)
	a.frontend.RunMainLoop()
	a.frontend.WindowClose()

	// Execute post-build hooks
	if a.pluginManager != nil {
		if buildCtx, ok := a.ctx.Value("buildCtx").(*plugins.BuildContext); ok {
			if hookErr := a.ExecuteBuildHook("PostBuild", buildCtx); hookErr != nil {
				a.logger.Error("Post-build hook failed: %v", hookErr)
				// Return the original error if there was one, otherwise the hook error
				if err == nil {
					err = hookErr
				}
			}
		}
	}

	if a.shutdownCallback != nil {
		a.shutdownCallback(a.ctx)
	}
	return err
}

// CreateApp creates the app!
func CreateApp(appoptions *options.App) (*App, error) {
	var err error

	ctx := context.Background()

	// Merge default options
	options.MergeDefaults(appoptions)

	debug := IsDebug()
	devtoolsEnabled := IsDevtoolsEnabled()
	ctx = context.WithValue(ctx, "debug", debug)
	ctx = context.WithValue(ctx, "devtoolsEnabled", devtoolsEnabled)

	// Set up logger
	myLogger := logger.New(appoptions.Logger)
	if IsDebug() {
		myLogger.SetLogLevel(appoptions.LogLevel)
	} else {
		myLogger.SetLogLevel(appoptions.LogLevelProduction)
	}
	ctx = context.WithValue(ctx, "logger", myLogger)
	ctx = context.WithValue(ctx, "obfuscated", IsObfuscated())

	// Preflight Checks
	err = PreflightChecks(appoptions, myLogger)
	if err != nil {
		return nil, err
	}

	// Create the menu manager
	menuManager := menumanager.NewManager()

	// Process the application menu
	if appoptions.Menu != nil {
		err = menuManager.SetApplicationMenu(appoptions.Menu)
		if err != nil {
			return nil, err
		}
	}

	// Create binding exemptions - Ugly hack. There must be a better way
	bindingExemptions := []interface{}{
		appoptions.OnStartup,
		appoptions.OnShutdown,
		appoptions.OnDomReady,
		appoptions.OnBeforeClose,
	}
	appBindings := binding.NewBindings(myLogger, appoptions.Bind, bindingExemptions, IsObfuscated(), appoptions.EnumBind)
	eventHandler := runtime.NewEvents(myLogger)
	ctx = context.WithValue(ctx, "events", eventHandler)
	// Attach logger to context
	if debug {
		ctx = context.WithValue(ctx, "buildtype", "debug")
	} else {
		ctx = context.WithValue(ctx, "buildtype", "production")
	}

	messageDispatcher := dispatcher.NewDispatcher(ctx, myLogger, appBindings, eventHandler, appoptions.ErrorFormatter, appoptions.DisablePanicRecovery)
	appFrontend := desktop.NewFrontend(ctx, appoptions, myLogger, appBindings, messageDispatcher)
	eventHandler.AddFrontend(appFrontend)

	ctx = context.WithValue(ctx, "frontend", appFrontend)
	result := &App{
		ctx:              ctx,
		frontend:         appFrontend,
		logger:           myLogger,
		menuManager:      menuManager,
		startupCallback:  appoptions.OnStartup,
		shutdownCallback: appoptions.OnShutdown,
		debug:            debug,
		devtoolsEnabled:  devtoolsEnabled,
		options:          appoptions,
	}

	// Initialize plugins if configured
	if err = result.InitializePlugins(appoptions); err != nil {
		myLogger.Error("Plugin initialization failed: %v", err)
		// Continue even if plugin initialization fails (unless required)
	}

	return result, nil

}
