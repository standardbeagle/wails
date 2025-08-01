# Task: Build Process Integration
**Generated from Master Planning**: 2024-01-15
**Context Package**: `/requests/wails-plugin-fork/context/`

## Task Sizing Assessment
**File Count**: 3-4 files
**Estimated Time**: 60 minutes
**Token Estimate**: 100k tokens
**Complexity Level**: 3 (Complex)
**Parallelization Benefit**: LOW - Core build process modification
**Atomicity Assessment**: ✅ ATOMIC - Complete build integration
**Boundary Analysis**: ✅ CLEAR - Focused on build command

## Persona Assignment
**Persona**: Software Engineer
**Expertise Required**: Wails build process, Go compilation, asset embedding
**Worktree**: `~/work/wrf/wails-plugin-fork/`

## Context Summary
**Risk Level**: HIGH - Modifies critical build path
**Integration Points**: Affects all Wails applications using plugins
**Architecture Pattern**: Hook injection into build pipeline
**Similar Reference**: Existing build steps in Wails

### Task Scope Boundaries
**MODIFY Zone** (Direct Changes):
```yaml
primary_files:
  - /v2/cmd/wails/internal/commands/build/build.go       # Main build logic
  - /v2/cmd/wails/internal/commands/build/plugin_hooks.go # Plugin hook implementation
  - /v2/internal/build/build.go                          # Build context enhancement
  - /v2/cmd/wails/internal/commands/build/build_test.go  # Tests
```

**REVIEW Zone** (Check for Impact):
```yaml
check_integration:
  - /v2/internal/binding/                # Binding generation timing
  - /v2/internal/frontend/               # Frontend build integration
```

**IGNORE Zone** (Do Not Touch):
```yaml
ignore_completely:
  - /v2/build/                           # Build assets
  - /v2/internal/frontend/desktop/       # Platform-specific builds
```

## Task Requirements
**Objective**: Integrate plugin hooks into Wails build process for code generation and build customization

**Success Criteria**:
- [ ] Pre-build hooks execute before Go compilation
- [ ] Code generation integrates with binding generation
- [ ] Post-build hooks execute after compilation
- [ ] Build fails gracefully on plugin errors
- [ ] No impact when plugins disabled
- [ ] Performance overhead <5% of build time

**Validation Commands**:
```bash
# Test normal build
wails build                              # Should work as before

# Test with plugins
wails build -plugins                     # Should execute plugin hooks

# Test code generation
wails plugin generate && wails build     # Generated code included

# Benchmark
time wails build                         # Compare with/without plugins
```

## Implementation Details

### 1. Build Command Enhancement (`build.go`)
```go
// Add to existing build command
type BuildOptions struct {
    // ... existing fields ...
    
    // Skip plugin execution
    SkipPlugins bool
}

// Update build function
func Build(options *BuildOptions) error {
    // Load project configuration
    project, err := LoadProject()
    if err != nil {
        return err
    }
    
    // Initialize plugin manager if enabled
    var pluginManager *plugins.Manager
    if !options.SkipPlugins && project.Config.Plugins != nil && project.Config.Plugins.Enabled {
        pluginManager, err = initializePluginManager(project.Config)
        if err != nil {
            logger.Warn("Plugin initialization failed", "error", err)
            // Continue without plugins unless required
            if project.Config.Plugins.Required {
                return fmt.Errorf("required plugins failed: %w", err)
            }
        }
    }
    
    // Create build context
    buildCtx := &plugins.BuildContext{
        Context:     context.Background(),
        ProjectRoot: project.Path,
        OutputDir:   options.OutputDir,
        BuildMode:   getBuildMode(options),
        Platform:    options.Platform,
        Arch:        options.Arch,
        Logger:      logger,
        StartTime:   time.Now(),
    }
    
    // Pre-build hooks
    if pluginManager != nil {
        logger.Info("Executing pre-build hooks...")
        if err := executePreBuildHooks(pluginManager, buildCtx); err != nil {
            return fmt.Errorf("pre-build hooks failed: %w", err)
        }
    }
    
    // ... existing build steps ...
    
    // Insert code generation before binding generation
    if pluginManager != nil {
        logger.Info("Running plugin code generation...")
        if err := executeCodeGeneration(pluginManager, project, buildCtx); err != nil {
            return fmt.Errorf("code generation failed: %w", err)
        }
    }
    
    // ... rest of build process ...
    
    // Post-build hooks
    if pluginManager != nil {
        logger.Info("Executing post-build hooks...")
        if err := executePostBuildHooks(pluginManager, buildCtx); err != nil {
            // Post-build errors are warnings
            logger.Warn("Post-build hooks had errors", "error", err)
        }
    }
    
    return nil
}
```

### 2. Plugin Hook Implementation (`plugin_hooks.go`)
```go
package build

import (
    "context"
    "fmt"
    "path/filepath"
    
    "github.com/wailsapp/wails/v2/pkg/plugins"
)

// executePreBuildHooks runs pre-build plugins
func executePreBuildHooks(manager *plugins.Manager, ctx *plugins.BuildContext) error {
    hooks := manager.GetPluginsByType("BuildHook")
    if len(hooks) == 0 {
        return nil
    }
    
    for _, plugin := range hooks {
        hook := plugin.(plugins.BuildHook)
        
        logger.Debug("Running pre-build hook", "plugin", plugin.Name())
        
        if err := hook.PreBuild(ctx); err != nil {
            return fmt.Errorf("plugin %s pre-build failed: %w", plugin.Name(), err)
        }
    }
    
    return nil
}

// executeCodeGeneration runs code generation plugins
func executeCodeGeneration(manager *plugins.Manager, project *Project, buildCtx *plugins.BuildContext) error {
    generators := manager.GetPluginsByType("CodeGenerator")
    if len(generators) == 0 {
        return nil
    }
    
    genCtx := &plugins.GenerationContext{
        Context:       buildCtx.Context,
        ProjectRoot:   project.Path,
        OutputDir:     project.Path,
        ProjectConfig: extractProjectConfig(project),
        Logger:        buildCtx.Logger,
    }
    
    for _, plugin := range generators {
        gen := plugin.(plugins.CodeGenerator)
        
        logger.Debug("Running code generation", "plugin", plugin.Name())
        
        files, err := gen.GenerateCode(genCtx)
        if err != nil {
            return fmt.Errorf("plugin %s generation failed: %w", plugin.Name(), err)
        }
        
        // Write generated files
        for _, file := range files {
            if err := writeGeneratedFile(project.Path, file); err != nil {
                return fmt.Errorf("failed to write generated file %s: %w", file.Path, err)
            }
        }
        
        logger.Info("Generated files", "plugin", plugin.Name(), "count", len(files))
    }
    
    return nil
}

// executePostBuildHooks runs post-build plugins
func executePostBuildHooks(manager *plugins.Manager, ctx *plugins.BuildContext) error {
    hooks := manager.GetPluginsByType("BuildHook")
    if len(hooks) == 0 {
        return nil
    }
    
    var errors []error
    
    for _, plugin := range hooks {
        hook := plugin.(plugins.BuildHook)
        
        logger.Debug("Running post-build hook", "plugin", plugin.Name())
        
        if err := hook.PostBuild(ctx); err != nil {
            errors = append(errors, fmt.Errorf("plugin %s: %w", plugin.Name(), err))
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("post-build errors: %v", errors)
    }
    
    return nil
}

// writeGeneratedFile safely writes a generated file
func writeGeneratedFile(projectRoot string, file *plugins.GeneratedFile) error {
    fullPath := filepath.Join(projectRoot, file.Path)
    
    // Ensure directory exists
    dir := filepath.Dir(fullPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }
    
    // Check if file exists and overwrite flag
    if !file.Overwrite {
        if _, err := os.Stat(fullPath); err == nil {
            logger.Debug("Skipping existing file", "path", file.Path)
            return nil
        }
    }
    
    // Write file with proper permissions
    mode := os.FileMode(file.Mode)
    if mode == 0 {
        mode = 0644
    }
    
    return os.WriteFile(fullPath, file.Content, mode)
}
```

### 3. Build Context Enhancement (`build/build.go`)
```go
// Add plugin-specific build information
type BuildInfo struct {
    // ... existing fields ...
    
    // Plugin information
    PluginsEnabled bool
    PluginList     []string
    GeneratedFiles []string
}

// Track generated files for cleanup
func (b *Builder) trackGeneratedFile(path string) {
    b.buildInfo.GeneratedFiles = append(b.buildInfo.GeneratedFiles, path)
}
```

## Error Handling Strategy
- Pre-build errors fail the build
- Code generation errors fail the build
- Post-build errors are warnings only
- Clear attribution of errors to specific plugins
- Timeout protection for long-running plugins

## Performance Optimization
- Parallel execution of independent plugins
- Caching of generated files when inputs unchanged
- Skip plugin initialization when disabled
- Benchmark critical paths

## Testing Requirements
- Test build with various plugin configurations
- Test error scenarios (plugin failures)
- Test generated file handling
- Verify no regression in normal builds
- Performance benchmarks

## Next Steps
With build integration complete, we can focus on the development server enhancement.