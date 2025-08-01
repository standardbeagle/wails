# Task: Development Server Plugin Hooks
**Generated from Master Planning**: 2024-01-15
**Context Package**: `/requests/wails-plugin-fork/context/`

## Task Sizing Assessment
**File Count**: 4-5 files
**Estimated Time**: 60 minutes
**Token Estimate**: 100k tokens
**Complexity Level**: 3 (Complex)
**Parallelization Benefit**: LOW - Core dev server modification
**Atomicity Assessment**: ✅ ATOMIC - Complete dev server integration
**Boundary Analysis**: ✅ CLEAR - Focused on dev command

## Persona Assignment
**Persona**: Software Engineer
**Expertise Required**: HTTP servers, WebSocket, file watching, hot reload
**Worktree**: `~/work/wrf/wails-plugin-fork/`

## Context Summary
**Risk Level**: HIGH - Modifies development workflow
**Integration Points**: Affects developer experience significantly
**Architecture Pattern**: Middleware chain, WebSocket multiplexing
**Similar Reference**: Existing Wails dev server

### Task Scope Boundaries
**MODIFY Zone** (Direct Changes):
```yaml
primary_files:
  - /v2/cmd/wails/internal/commands/dev/dev.go           # Main dev command
  - /v2/cmd/wails/internal/commands/dev/plugin_server.go # Plugin server extensions
  - /v2/cmd/wails/internal/commands/dev/watcher.go       # File watching integration
  - /v2/cmd/wails/internal/commands/dev/websocket.go     # WebSocket enhancements
```

**REVIEW Zone** (Check for Impact):
```yaml
check_integration:
  - /v2/internal/frontend/devserver/     # Dev server implementation
  - /v2/pkg/assetserver/                 # Asset serving patterns
```

**IGNORE Zone** (Do Not Touch):
```yaml
ignore_completely:
  - /v2/internal/frontend/desktop/       # Platform-specific code
  - /v2/internal/binding/                # Binding generation (separate)
```

## Task Requirements
**Objective**: Enable plugins to extend development server with middleware, WebSocket handlers, and file watching

**Success Criteria**:
- [ ] Plugins can add HTTP middleware
- [ ] Plugins can handle WebSocket connections
- [ ] Plugins can register file watchers
- [ ] Plugin dev tools accessible via server
- [ ] Hot reload works with plugin changes
- [ ] No impact when plugins disabled

**Validation Commands**:
```bash
# Test normal dev mode
wails dev                                # Should work as before

# Test with plugins
wails dev                                # Should load plugin extensions

# Access plugin tools
curl http://localhost:34115/plugin/wrf/info  # Plugin endpoints work

# Test hot reload
# Edit Go file → should trigger plugin regeneration
```

## Implementation Details

### 1. Dev Command Enhancement (`dev.go`)
```go
// Add plugin support to dev command
func Dev(options *DevOptions) error {
    // ... existing setup ...
    
    // Initialize plugin manager if enabled
    var pluginManager *plugins.Manager
    if project.Config.Plugins != nil && project.Config.Plugins.Enabled {
        pluginManager, err = initializePluginManager(project.Config)
        if err != nil {
            logger.Warn("Plugin initialization failed", "error", err)
        } else {
            defer pluginManager.Shutdown(context.Background())
        }
    }
    
    // Create enhanced dev server
    server := &DevServer{
        options:       options,
        pluginManager: pluginManager,
        logger:        logger,
    }
    
    // Configure plugin extensions
    if pluginManager != nil {
        if err := configurePluginExtensions(server, pluginManager); err != nil {
            logger.Warn("Plugin configuration failed", "error", err)
        }
    }
    
    // Start enhanced file watching
    if err := server.startWatching(); err != nil {
        return err
    }
    
    // Start dev server with plugin support
    return server.Start()
}
```

### 2. Plugin Server Extensions (`plugin_server.go`)
```go
package dev

import (
    "net/http"
    "github.com/gorilla/mux"
    "github.com/gorilla/websocket"
    "github.com/wailsapp/wails/v2/pkg/plugins"
)

// DevServer with plugin support
type DevServer struct {
    // ... existing fields ...
    
    pluginManager  *plugins.Manager
    pluginHandlers map[string]http.Handler
    wsHandlers     map[string]plugins.WebSocketHandler
    middleware     []plugins.Middleware
}

// configurePluginExtensions sets up plugin extensions
func configurePluginExtensions(server *DevServer, manager *plugins.Manager) error {
    extensions := manager.GetPluginsByType("DevServerExtension")
    
    for _, plugin := range extensions {
        ext := plugin.(plugins.DevServerExtension)
        
        // Create configurator for this plugin
        configurator := &devServerConfigurator{
            server:     server,
            pluginName: plugin.Name(),
        }
        
        if err := ext.ConfigureDevServer(configurator); err != nil {
            return fmt.Errorf("plugin %s configuration failed: %w", plugin.Name(), err)
        }
    }
    
    return nil
}

// devServerConfigurator implements plugins.DevServerConfigurator
type devServerConfigurator struct {
    server     *DevServer
    pluginName string
}

// AddMiddleware adds HTTP middleware
func (c *devServerConfigurator) AddMiddleware(middleware plugins.Middleware) {
    c.server.middleware = append(c.server.middleware, middleware)
}

// AddHandler adds HTTP handler for plugin routes
func (c *devServerConfigurator) AddHandler(path string, handler http.Handler) {
    // Namespace plugin routes
    fullPath := fmt.Sprintf("/plugin/%s%s", c.pluginName, path)
    c.server.pluginHandlers[fullPath] = handler
}

// AddWebSocketHandler adds WebSocket handler
func (c *devServerConfigurator) AddWebSocketHandler(path string, handler plugins.WebSocketHandler) {
    fullPath := fmt.Sprintf("/plugin/%s%s", c.pluginName, path)
    c.server.wsHandlers[fullPath] = handler
}

// Enhanced HTTP handler with plugin support
func (s *DevServer) createHTTPHandler() http.Handler {
    router := mux.NewRouter()
    
    // Apply plugin middleware
    handler := http.Handler(router)
    for i := len(s.middleware) - 1; i >= 0; i-- {
        handler = s.middleware[i](handler)
    }
    
    // Register plugin routes
    for path, h := range s.pluginHandlers {
        router.PathPrefix(path).Handler(h)
    }
    
    // WebSocket upgrade handling
    router.HandleFunc("/ws", s.handleWebSocket)
    
    // Plugin-specific WebSocket routes
    for path, wsHandler := range s.wsHandlers {
        router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
            s.handlePluginWebSocket(w, r, wsHandler)
        })
    }
    
    // ... existing routes ...
    
    return handler
}
```

### 3. Enhanced File Watching (`watcher.go`)
```go
// startWatching with plugin support
func (s *DevServer) startWatching() error {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return err
    }
    
    s.watcher = watcher
    
    // Add standard paths
    // ... existing watch paths ...
    
    // Add plugin watch paths
    if s.pluginManager != nil {
        watchers := s.pluginManager.GetPluginsByType("FileWatcher")
        
        for _, plugin := range watchers {
            fw := plugin.(plugins.FileWatcher)
            paths, err := fw.GetWatchPaths(s.options.ProjectRoot)
            if err != nil {
                s.logger.Warn("Failed to get watch paths", "plugin", plugin.Name(), "error", err)
                continue
            }
            
            for _, path := range paths {
                if err := watcher.Add(path); err != nil {
                    s.logger.Warn("Failed to watch path", "path", path, "error", err)
                } else {
                    s.logger.Debug("Watching path", "path", path, "plugin", plugin.Name())
                }
            }
        }
    }
    
    // Start watching
    go s.watchLoop()
    
    return nil
}

// watchLoop with plugin notifications
func (s *DevServer) watchLoop() {
    for {
        select {
        case event, ok := <-s.watcher.Events:
            if !ok {
                return
            }
            
            // Notify plugins of file changes
            if s.pluginManager != nil {
                ctx := &plugins.FileChangeContext{
                    Context:     context.Background(),
                    FilePath:    event.Name,
                    EventType:   getEventType(event),
                    ProjectRoot: s.options.ProjectRoot,
                    Logger:      s.logger,
                }
                
                s.pluginManager.ExecuteHook("OnFileChanged", ctx)
            }
            
            // ... existing file change handling ...
            
        case err, ok := <-s.watcher.Errors:
            if !ok {
                return
            }
            s.logger.Error("Watcher error", "error", err)
        }
    }
}
```

### 4. WebSocket Enhancement (`websocket.go`)
```go
// WebSocket multiplexing for plugins
type wsHub struct {
    clients    map[*wsClient]bool
    broadcast  chan []byte
    register   chan *wsClient
    unregister chan *wsClient
    
    // Plugin channels
    pluginChannels map[string]chan []byte
}

// handlePluginWebSocket handles plugin-specific WebSocket connections
func (s *DevServer) handlePluginWebSocket(w http.ResponseWriter, r *http.Request, handler plugins.WebSocketHandler) {
    upgrader := websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool {
            return true // Dev mode allows all origins
        },
    }
    
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        s.logger.Error("WebSocket upgrade failed", "error", err)
        return
    }
    defer conn.Close()
    
    // Create plugin WebSocket context
    ctx := &plugins.WebSocketContext{
        Conn:        conn,
        Request:     r,
        Logger:      s.logger,
        ProjectRoot: s.options.ProjectRoot,
    }
    
    // Let plugin handle the connection
    if err := handler.HandleWebSocket(ctx); err != nil {
        s.logger.Error("Plugin WebSocket handler error", "error", err)
    }
}

// Broadcast to specific plugin channel
func (s *DevServer) broadcastToPlugin(pluginName string, message []byte) {
    if ch, ok := s.hub.pluginChannels[pluginName]; ok {
        select {
        case ch <- message:
        default:
            // Channel full, drop message
            s.logger.Warn("Plugin channel full", "plugin", pluginName)
        }
    }
}
```

## Plugin Examples

### React DevTools Plugin
```go
func (p *ReactDevToolsPlugin) ConfigureDevServer(config plugins.DevServerConfigurator) error {
    // Add middleware for React DevTools
    config.AddMiddleware(func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("X-React-DevTools", "enabled")
            next.ServeHTTP(w, r)
        })
    })
    
    // Add handler for DevTools UI
    config.AddHandler("/devtools", http.HandlerFunc(p.serveDevTools))
    
    // Add WebSocket for DevTools communication
    config.AddWebSocketHandler("/devtools/ws", p)
    
    return nil
}
```

## Testing Strategy
- Test dev server with plugins enabled/disabled
- Test middleware ordering
- Test WebSocket multiplexing
- Test file watching with plugin paths
- Test hot reload with plugin changes

## Performance Considerations
- Lazy loading of plugin resources
- Efficient file watching (avoid duplicates)
- WebSocket message throttling
- Middleware performance impact

## Next Steps
With dev server hooks in place, we can implement the basic React plugin.