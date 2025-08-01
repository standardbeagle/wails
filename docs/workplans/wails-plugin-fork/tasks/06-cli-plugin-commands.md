# Task: CLI Plugin Commands
**Generated from Master Planning**: 2024-01-15
**Context Package**: `/requests/wails-plugin-fork/context/`

## Task Sizing Assessment
**File Count**: 4-5 files
**Estimated Time**: 45 minutes
**Token Estimate**: 70k tokens
**Complexity Level**: 2 (Moderate)
**Parallelization Benefit**: HIGH - Can be done parallel to other CLI work
**Atomicity Assessment**: ✅ ATOMIC - Standalone CLI feature
**Boundary Analysis**: ✅ CLEAR - New command files

## Persona Assignment
**Persona**: Software Engineer
**Expertise Required**: Cobra CLI framework, command patterns
**Worktree**: `~/work/wrf/wails-plugin-fork/`

## Context Summary
**Risk Level**: LOW - New commands, no breaking changes
**Integration Points**: Uses plugin manager, affects developer workflow
**Architecture Pattern**: Cobra command pattern
**Similar Reference**: Existing Wails CLI commands

### Task Scope Boundaries
**MODIFY Zone** (Direct Changes):
```yaml
primary_files:
  - /v2/cmd/wails/main.go                # Register plugin command
  - /v2/cmd/wails/internal/commands/plugin/plugin.go     # Main plugin command
  - /v2/cmd/wails/internal/commands/plugin/list.go       # List plugins
  - /v2/cmd/wails/internal/commands/plugin/info.go       # Plugin info
  - /v2/cmd/wails/internal/commands/plugin/generate.go   # Trigger generation
```

**REVIEW Zone** (Check for Impact):
```yaml
check_patterns:
  - /v2/cmd/wails/internal/commands/     # Command patterns
  - /v2/internal/project/                # Project detection
```

**IGNORE Zone** (Do Not Touch):
```yaml
ignore_completely:
  - /v2/cmd/wails/internal/commands/build/  # Build command (separate task)
  - /v2/cmd/wails/internal/commands/dev/    # Dev command (separate task)
```

## Task Requirements
**Objective**: Implement CLI commands for plugin management and interaction

**Success Criteria**:
- [ ] `wails plugin` command group implemented
- [ ] `wails plugin list` shows available plugins
- [ ] `wails plugin info <name>` shows plugin details
- [ ] `wails plugin generate` triggers code generation
- [ ] Commands work with and without wails.json
- [ ] Helpful error messages and usage docs

**Validation Commands**:
```bash
wails plugin --help                      # Shows help
wails plugin list                        # Lists plugins
wails plugin info wrf                    # Shows WRF plugin info
wails plugin generate                    # Runs code generation
```

## Implementation Details

### 1. Main Command Registration (`main.go`)
```go
// Add to existing command tree
func init() {
    // ... existing commands ...
    
    rootCmd.AddCommand(pluginCmd)
}
```

### 2. Plugin Root Command (`plugin/plugin.go`)
```go
package plugin

import (
    "github.com/spf13/cobra"
    "github.com/wailsapp/wails/v2/cmd/wails/internal/commands"
    "github.com/wailsapp/wails/v2/pkg/plugins"
)

var pluginCmd = &cobra.Command{
    Use:   "plugin",
    Short: "Plugin management commands",
    Long: `Manage Wails plugins for enhanced development experience.

Plugins extend Wails with:
  - Code generation from Go to TypeScript
  - Framework-specific integrations
  - Development tools and debugging
  - Custom build steps`,
    Example: `  # List available plugins
  wails plugin list
  
  # Get information about a plugin
  wails plugin info wrf
  
  # Run code generation
  wails plugin generate`,
}

func init() {
    pluginCmd.AddCommand(listCmd)
    pluginCmd.AddCommand(infoCmd)
    pluginCmd.AddCommand(generateCmd)
}
```

### 3. List Command (`list.go`)
```go
var listCmd = &cobra.Command{
    Use:   "list",
    Short: "List available plugins",
    Long:  "List all available plugins including built-in, project, and user plugins",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Load configuration
        project := commands.FindProject()
        config := loadPluginConfig(project)
        
        // Create temporary plugin manager for listing
        manager := plugins.NewManager(&plugins.ManagerConfig{
            DiscoveryPaths: config.DiscoveryPaths,
            Logger:         commands.NewLogger(),
        })
        
        // Discover plugins without initializing
        plugins := manager.DiscoverPlugins()
        
        // Display plugins
        if len(plugins) == 0 {
            fmt.Println("No plugins found.")
            return nil
        }
        
        fmt.Println("Available Wails Plugins:")
        fmt.Println()
        
        // Group by source
        for source, pluginList := range groupPluginsBySource(plugins) {
            fmt.Printf("%s:\n", source)
            for _, p := range pluginList {
                status := "available"
                if isEnabled(p.Name(), config) {
                    status = "enabled"
                }
                fmt.Printf("  - %s (v%s) - %s [%s]\n", 
                    p.Name(), p.Version(), p.Description(), status)
            }
            fmt.Println()
        }
        
        return nil
    },
}
```

### 4. Info Command (`info.go`)
```go
var infoCmd = &cobra.Command{
    Use:   "info [plugin-name]",
    Short: "Show detailed information about a plugin",
    Long:  "Display detailed information about a specific plugin including configuration options",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        pluginName := args[0]
        
        // Similar setup to list
        project := commands.FindProject()
        config := loadPluginConfig(project)
        manager := createPluginManager(config)
        
        // Find plugin
        plugin, err := manager.FindPlugin(pluginName)
        if err != nil {
            return fmt.Errorf("plugin '%s' not found", pluginName)
        }
        
        // Display detailed info
        fmt.Printf("Plugin: %s\n", plugin.Name())
        fmt.Printf("Version: %s\n", plugin.Version())
        fmt.Printf("Author: %s\n", plugin.Author())
        fmt.Printf("Description: %s\n", plugin.Description())
        fmt.Println()
        
        // Show capabilities
        fmt.Println("Capabilities:")
        if _, ok := plugin.(plugins.CodeGenerator); ok {
            fmt.Println("  - Code Generation")
        }
        if _, ok := plugin.(plugins.BuildHook); ok {
            fmt.Println("  - Build Integration")
        }
        if _, ok := plugin.(plugins.DevServerExtension); ok {
            fmt.Println("  - Development Server Extension")
        }
        if _, ok := plugin.(plugins.FileWatcher); ok {
            fmt.Println("  - File Watching")
        }
        if _, ok := plugin.(plugins.TemplateProvider); ok {
            fmt.Println("  - Project Templates")
        }
        
        // Configuration example
        if example := getConfigExample(pluginName); example != "" {
            fmt.Println("\nConfiguration Example:")
            fmt.Println(example)
        }
        
        return nil
    },
}
```

### 5. Generate Command (`generate.go`)
```go
var generateCmd = &cobra.Command{
    Use:   "generate",
    Short: "Run code generation for enabled plugins",
    Long:  "Execute code generation for all enabled plugins that support it",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Require project context
        project := commands.FindProject()
        if project == nil {
            return fmt.Errorf("no Wails project found")
        }
        
        // Load full configuration
        appConfig, err := commands.LoadAppConfig(project)
        if err != nil {
            return err
        }
        
        if appConfig.Plugins == nil || !appConfig.Plugins.Enabled {
            return fmt.Errorf("plugins not enabled in wails.json")
        }
        
        // Initialize plugin manager
        manager := plugins.NewManager(&plugins.ManagerConfig{
            Plugins:        appConfig.Plugins.Plugins,
            DiscoveryPaths: appConfig.Plugins.DiscoveryPaths,
            Logger:         commands.NewLogger(),
        })
        
        ctx := context.Background()
        if err := manager.Initialize(ctx, appConfig); err != nil {
            return fmt.Errorf("plugin initialization failed: %w", err)
        }
        defer manager.Shutdown(ctx)
        
        // Execute code generation
        genCtx := &plugins.GenerationContext{
            Context:       ctx,
            ProjectRoot:   project.Path,
            OutputDir:     project.Path,
            ProjectConfig: extractProjectConfig(appConfig),
            Logger:        commands.NewLogger(),
        }
        
        fmt.Println("Running code generation...")
        if err := manager.ExecuteHook("GenerateCode", genCtx); err != nil {
            return fmt.Errorf("code generation failed: %w", err)
        }
        
        fmt.Println("Code generation completed successfully!")
        return nil
    },
}
```

## Helper Functions
- `loadPluginConfig()` - Load plugin configuration with defaults
- `groupPluginsBySource()` - Organize plugins by built-in/project/user
- `isEnabled()` - Check if plugin is enabled in config
- `getConfigExample()` - Return configuration example for plugin

## Error Handling
- Clear messages when no project found
- Helpful errors for missing plugins
- Validate configuration before use

## Next Steps
With CLI commands in place, we can implement the build process integration.