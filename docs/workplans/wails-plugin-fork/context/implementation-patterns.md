# Implementation Patterns - Wails Plugin Fork

## Code Patterns to Follow

### 1. Error Handling Pattern (from Wails)
```go
// Follow Wails error wrapping pattern
if err != nil {
    return fmt.Errorf("plugin %s initialization failed: %w", name, err)
}

// Use structured logging
app.logger.Error("Plugin error", "plugin", name, "error", err)
```

### 2. Configuration Pattern (from pkg/options)
```go
// Extend existing options pattern
type PluginConfig struct {
    Enabled        bool                       `json:"enabled"`
    Plugins        map[string]*PluginEntry   `json:"plugins,omitempty"`
    DiscoveryPaths []string                  `json:"discoveryPaths,omitempty"`
}

// Merge with existing App options
type App struct {
    // ... existing fields ...
    Plugins *PluginConfig `json:"plugins,omitempty"`
}
```

### 3. Command Pattern (from cmd/wails)
```go
// Follow Cobra command structure
var pluginCmd = &cobra.Command{
    Use:   "plugin",
    Short: "Plugin management commands",
    Long:  `Manage Wails plugins for enhanced development experience`,
}

func init() {
    rootCmd.AddCommand(pluginCmd)
    pluginCmd.AddCommand(pluginListCmd)
    pluginCmd.AddCommand(pluginInfoCmd)
}
```

### 4. Middleware Pattern (from assetserver)
```go
// Chain middleware following existing pattern
type Middleware func(http.Handler) http.Handler

func (s *Server) Use(middleware ...Middleware) {
    s.middleware = append(s.middleware, middleware...)
}

// Plugin middleware registration
func (p *DevServerPlugin) Middleware() Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Plugin logic
            next.ServeHTTP(w, r)
        })
    }
}
```

### 5. File Watching Pattern (from dev command)
```go
// Follow existing watcher pattern
watcher, err := fsnotify.NewWatcher()
if err != nil {
    return err
}
defer watcher.Close()

// Plugin file watching
for _, plugin := range plugins {
    if fw, ok := plugin.(FileWatcher); ok {
        paths, _ := fw.GetWatchPaths(projectRoot)
        for _, path := range paths {
            watcher.Add(path)
        }
    }
}
```

### 6. Template Pattern (from pkg/templates)
```go
// Follow template registration pattern
type PluginTemplate struct {
    Name        string
    Description string
    Path        string
}

// Register templates from plugins
func (m *Manager) RegisterTemplates() {
    for _, plugin := range m.plugins {
        if tp, ok := plugin.(TemplateProvider); ok {
            templates := tp.GetTemplates()
            // Register with template system
        }
    }
}
```

## Naming Conventions

### Package Names
- `plugins` - Main plugin package
- `pluginutil` - Plugin utilities
- `plugingen` - Code generation utilities

### Interface Names
- `Plugin` - Base interface
- `BuildHook` - Build integration
- `CodeGenerator` - Code generation
- `DevServerExtension` - Dev server integration

### File Names
- `interface.go` - Core interfaces
- `manager.go` - Plugin manager
- `context.go` - Context types
- `builtin.go` - Built-in plugins

## Testing Patterns

### Mock Plugin Pattern
```go
type MockPlugin struct {
    name    string
    version string
    mock.Mock
}

func (m *MockPlugin) Name() string { return m.name }
func (m *MockPlugin) Version() string { return m.version }

func TestPluginManager(t *testing.T) {
    mockPlugin := &MockPlugin{name: "test", version: "1.0.0"}
    // Test plugin lifecycle
}
```

### Integration Test Pattern
```go
func TestPluginIntegration(t *testing.T) {
    // Create test project
    tmpDir := t.TempDir()
    
    // Initialize plugin system
    manager := NewManager(config, logger)
    
    // Test full lifecycle
    err := manager.Initialize(ctx, wailsConfig)
    require.NoError(t, err)
    
    // Verify behavior
    // ...
    
    err = manager.Shutdown(ctx)
    require.NoError(t, err)
}
```

## Code Generation Patterns

### Annotation Pattern
```go
// Support annotations like existing Wails bindings
// @wrf:query(cache: "5m")
func (s *Service) GetUser(id string) (*User, error)

// Parse annotations during AST traversal
func parseAnnotations(comments []*ast.CommentGroup) map[string]string {
    annotations := make(map[string]string)
    // Parse @wrf: prefixed comments
    return annotations
}
```

### Template Generation Pattern
```go
// Use templates for code generation
const hookTemplate = `
export function {{ .HookName }}({{ .Params }}) {
    return useQuery({
        queryKey: [{{ .QueryKey }}],
        queryFn: () => window.go.{{ .Service }}.{{ .Method }}({{ .Args }}),
        {{ .Options }}
    })
}
`

// Generate using templates
tmpl := template.Must(template.New("hook").Parse(hookTemplate))
err := tmpl.Execute(output, data)
```

## Documentation Patterns

### GoDoc Comments
```go
// Package plugins provides the Wails plugin system for extending
// build, development, and code generation capabilities.
//
// The plugin system is designed to be minimally invasive to core Wails
// while providing powerful extension points for framework integration.
package plugins

// Plugin defines the core interface that all Wails plugins must implement.
// Plugins can optionally implement additional interfaces for extended functionality.
type Plugin interface {
    // Name returns the unique name of the plugin.
    Name() string
    
    // Version returns the semantic version of the plugin.
    Version() string
}
```