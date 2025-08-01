# Task: Application Initialization Hooks
**Generated from Master Planning**: 2024-01-15
**Context Package**: `/requests/wails-plugin-fork/context/`

## Task Sizing Assessment
**File Count**: 2-3 files
**Estimated Time**: 45 minutes
**Token Estimate**: 60k tokens
**Complexity Level**: 3 (Complex)
**Parallelization Benefit**: LOW - Core integration point
**Atomicity Assessment**: ✅ ATOMIC - Critical integration
**Boundary Analysis**: ✅ CLEAR - Minimal changes to existing code

## Persona Assignment
**Persona**: System Architect
**Expertise Required**: Wails internals, dependency injection, initialization flow
**Worktree**: `~/work/wrf/wails-plugin-fork/`

## Context Summary
**Risk Level**: HIGH - Modifies core Wails initialization
**Integration Points**: All plugins depend on proper initialization
**Architecture Pattern**: Hook injection with minimal changes
**Similar Reference**: Existing Wails service initialization

### Task Scope Boundaries
**MODIFY Zone** (Direct Changes):
```yaml
primary_files:
  - /v2/internal/app/app.go              # Add plugin manager initialization
  - /v2/internal/app/app_dev.go          # Dev mode plugin hooks
  - /v2/internal/app/app_production.go   # Production mode plugin hooks
```

**REVIEW Zone** (Check for Impact):
```yaml
check_initialization:
  - /v2/pkg/wails/wails.go              # Main entry point
  - /v2/internal/frontend/               # Frontend initialization
```

**IGNORE Zone** (Do Not Touch):
```yaml
ignore_completely:
  - /v2/internal/binding/                # Binding generation (later phase)
  - /v2/internal/frontend/desktop/       # Platform-specific code
```

## Task Requirements
**Objective**: Integrate plugin system into Wails application lifecycle with minimal changes

**Success Criteria**:
- [ ] Plugin manager initialized during app startup
- [ ] Plugins loaded based on configuration
- [ ] Proper error handling without breaking existing apps
- [ ] Dev and production mode support
- [ ] Clean shutdown of plugins
- [ ] Zero impact when plugins disabled

**Validation Commands**:
```bash
# Test with plugins disabled
go run ./cmd/wails build                 # Should work as before

# Test with plugins enabled
echo '{"plugins":{"enabled":true}}' > wails.json
go run ./cmd/wails build                 # Should initialize plugins

# Run tests
go test ./v2/internal/app/...
```

## Implementation Details

### 1. App Structure Enhancement (`app.go`)
```go
// Add to App struct
type App struct {
    // ... existing fields ...
    
    // Plugin manager (optional, nil when disabled)
    pluginManager *plugins.Manager
}

// Update NewApp or initialization
func (a *App) Initialize(options *options.App) error {
    // ... existing initialization ...
    
    // Initialize plugin system if enabled
    if options.Plugins != nil && options.Plugins.Enabled {
        if err := a.initializePlugins(options); err != nil {
            // Log error but don't fail app startup
            a.logger.Error("Plugin system initialization failed", "error", err)
            // Optionally continue without plugins based on config
            if options.Plugins.Required {
                return fmt.Errorf("required plugin system failed: %w", err)
            }
        }
    }
    
    return nil
}

// Plugin initialization
func (a *App) initializePlugins(options *options.App) error {
    // Create plugin manager
    config := &plugins.ManagerConfig{
        Plugins:        options.Plugins.Plugins,
        DiscoveryPaths: options.Plugins.DiscoveryPaths,
        LoadTimeout:    options.Plugins.LoadTimeout,
        Logger:         a.logger.WithField("component", "plugins"),
    }
    
    a.pluginManager = plugins.NewManager(config)
    
    // Initialize plugins
    ctx, cancel := context.WithTimeout(context.Background(), config.LoadTimeout)
    defer cancel()
    
    return a.pluginManager.Initialize(ctx, options)
}

// Update Shutdown
func (a *App) Shutdown() {
    // ... existing shutdown ...
    
    // Shutdown plugins
    if a.pluginManager != nil {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        
        if err := a.pluginManager.Shutdown(ctx); err != nil {
            a.logger.Error("Plugin shutdown error", "error", err)
        }
    }
}
```

### 2. Dev Mode Integration (`app_dev.go`)
```go
// Hook into development mode
func (a *App) RunDev() error {
    // Execute pre-dev hooks
    if a.pluginManager != nil {
        devCtx := &plugins.DevContext{
            Context:     context.Background(),
            ProjectRoot: a.options.ProjectRoot,
            ServerPort:  a.options.DevServer.Port,
            Logger:      a.logger,
            StartTime:   time.Now(),
        }
        
        if err := a.pluginManager.ExecuteHook("PreDev", devCtx); err != nil {
            a.logger.Warn("Pre-dev hook errors", "error", err)
        }
    }
    
    // ... existing dev mode code ...
    
    // Post-dev hooks
    if a.pluginManager != nil {
        if err := a.pluginManager.ExecuteHook("PostDev", devCtx); err != nil {
            a.logger.Warn("Post-dev hook errors", "error", err)
        }
    }
}
```

### 3. Production Mode Integration (`app_production.go`)
```go
// Similar pattern for production builds
func (a *App) Build() error {
    // Pre-build hooks
    if a.pluginManager != nil {
        buildCtx := &plugins.BuildContext{
            Context:     context.Background(),
            ProjectRoot: a.options.ProjectRoot,
            OutputDir:   a.options.Build.OutputDir,
            BuildMode:   "production",
            Platform:    runtime.GOOS,
            Arch:        runtime.GOARCH,
            Logger:      a.logger,
            StartTime:   time.Now(),
        }
        
        if err := a.pluginManager.ExecuteHook("PreBuild", buildCtx); err != nil {
            return fmt.Errorf("pre-build hook failed: %w", err)
        }
    }
    
    // ... existing build code ...
    
    // Post-build hooks
    if a.pluginManager != nil {
        if err := a.pluginManager.ExecuteHook("PostBuild", buildCtx); err != nil {
            a.logger.Warn("Post-build hook errors", "error", err)
        }
    }
}
```

## Error Handling Strategy
- Plugin errors logged but don't crash app (unless configured as required)
- Timeouts on all plugin operations
- Panic recovery in plugin execution
- Clear error messages indicating plugin source

## Testing Strategy
- Test with plugins enabled/disabled
- Test plugin initialization failure scenarios
- Test timeout handling
- Verify no regression in existing functionality

## Performance Considerations
- Lazy loading of plugin manager
- No overhead when plugins disabled
- Timeout limits on plugin operations

## Next Steps
With initialization hooks in place, we can implement the CLI commands for plugin management.