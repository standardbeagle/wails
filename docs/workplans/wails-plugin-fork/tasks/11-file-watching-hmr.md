# Task: File Watching and Hot Module Replacement
**Generated from Master Planning**: 2024-01-15
**Context Package**: `/requests/wails-plugin-fork/context/`

## Task Sizing Assessment
**File Count**: 4-5 files
**Estimated Time**: 45 minutes
**Token Estimate**: 80k tokens
**Complexity Level**: 3 (Complex)
**Parallelization Benefit**: LOW - Integrates with dev server
**Atomicity Assessment**: ✅ ATOMIC - Complete HMR system
**Boundary Analysis**: ✅ CLEAR - Focused on watching/HMR

## Persona Assignment
**Persona**: Software Engineer
**Expertise Required**: File watching, WebSocket, HMR protocols
**Worktree**: `~/work/wrf/wails-plugin-fork/`

## Context Summary
**Risk Level**: MEDIUM - Affects development experience
**Integration Points**: Dev server, plugins, frontend build
**Architecture Pattern**: Event-driven file watching with WebSocket
**Similar Reference**: Existing Wails file watching

### Task Scope Boundaries
**MODIFY Zone** (Direct Changes):
```yaml
primary_files:
  - /v2/cmd/wails/internal/commands/dev/filewatcher.go    # Enhanced file watcher
  - /v2/cmd/wails/internal/commands/dev/hmr.go            # HMR protocol
  - /v2/pkg/plugins/wrf/watcher.go                        # WRF file watching
  - /v2/pkg/plugins/wrf/hmr.go                            # WRF HMR logic
  - /v2/internal/frontend/devserver/hmr.go                # HMR WebSocket
```

**REVIEW Zone** (Check for Impact):
```yaml
check_integration:
  - /v2/cmd/wails/internal/commands/dev/dev.go    # Dev server integration
  - /v2/internal/frontend/devserver/             # Frontend server
```

**IGNORE Zone** (Do Not Touch):
```yaml
ignore_completely:
  - /v2/internal/binding/                        # Binding generation
  - /v2/build/                                   # Build assets
```

## Task Requirements
**Objective**: Implement intelligent file watching and HMR for plugin-generated code

**Success Criteria**:
- [ ] Watch Go files and trigger plugin regeneration
- [ ] Smart HMR that preserves React state
- [ ] Debounced file change handling
- [ ] WebSocket protocol for HMR messages
- [ ] Selective component updates
- [ ] Graceful error handling

**Validation Commands**:
```bash
# Start dev mode with watching
wails dev

# Test Go file changes
echo "// test" >> app/service.go          # Should trigger regeneration

# Test HMR
# Edit frontend component → should hot reload without losing state

# Monitor WebSocket messages
wscat -c ws://localhost:34115/ws          # See HMR messages
```

## Implementation Details

### 1. Enhanced File Watcher (`filewatcher.go`)
```go
package dev

import (
    "path/filepath"
    "sync"
    "time"
    
    "github.com/fsnotify/fsnotify"
    "github.com/wailsapp/wails/v2/pkg/plugins"
)

// EnhancedWatcher wraps fsnotify with plugin support
type EnhancedWatcher struct {
    watcher       *fsnotify.Watcher
    pluginManager *plugins.Manager
    server        *DevServer
    logger        Logger
    
    // Debouncing
    debouncer     *Debouncer
    
    // File categorization
    goFiles       map[string]time.Time
    pluginFiles   map[string]string // file -> plugin name
    
    mu            sync.RWMutex
}

// NewEnhancedWatcher creates enhanced file watcher
func NewEnhancedWatcher(server *DevServer, pluginManager *plugins.Manager) (*EnhancedWatcher, error) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return nil, err
    }
    
    return &EnhancedWatcher{
        watcher:       watcher,
        pluginManager: pluginManager,
        server:        server,
        logger:        server.logger,
        debouncer:     NewDebouncer(500 * time.Millisecond),
        goFiles:       make(map[string]time.Time),
        pluginFiles:   make(map[string]string),
    }, nil
}

// Start watching with plugin integration
func (w *EnhancedWatcher) Start() error {
    // Add standard paths
    if err := w.addStandardPaths(); err != nil {
        return err
    }
    
    // Add plugin-specific paths
    if w.pluginManager != nil {
        if err := w.addPluginPaths(); err != nil {
            return err
        }
    }
    
    // Start event loop
    go w.eventLoop()
    
    return nil
}

// addPluginPaths adds paths from FileWatcher plugins
func (w *EnhancedWatcher) addPluginPaths() error {
    watchers := w.pluginManager.GetPluginsByType("FileWatcher")
    
    for _, plugin := range watchers {
        fw := plugin.(plugins.FileWatcher)
        paths, err := fw.GetWatchPaths(w.server.options.ProjectRoot)
        if err != nil {
            w.logger.Warn("Failed to get plugin paths", "plugin", plugin.Name(), "error", err)
            continue
        }
        
        for _, path := range paths {
            absPath, _ := filepath.Abs(path)
            if err := w.watcher.Add(absPath); err != nil {
                w.logger.Warn("Failed to watch", "path", path, "error", err)
            } else {
                // Track which plugin registered this path
                w.mu.Lock()
                w.pluginFiles[absPath] = plugin.Name()
                w.mu.Unlock()
                
                w.logger.Debug("Watching for plugin", "path", path, "plugin", plugin.Name())
            }
        }
    }
    
    return nil
}

// eventLoop processes file system events
func (w *EnhancedWatcher) eventLoop() {
    for {
        select {
        case event, ok := <-w.watcher.Events:
            if !ok {
                return
            }
            
            w.handleEvent(event)
            
        case err, ok := <-w.watcher.Errors:
            if !ok {
                return
            }
            w.logger.Error("Watch error", "error", err)
        }
    }
}

// handleEvent processes a file system event
func (w *EnhancedWatcher) handleEvent(event fsnotify.Event) {
    // Ignore certain files
    if shouldIgnore(event.Name) {
        return
    }
    
    // Debounce events
    w.debouncer.Debounce(event.Name, func() {
        w.processFileChange(event)
    })
}

// processFileChange handles debounced file changes
func (w *EnhancedWatcher) processFileChange(event fsnotify.Event) {
    filePath := event.Name
    ext := filepath.Ext(filePath)
    
    w.logger.Info("File changed", "file", filePath, "op", event.Op)
    
    // Categorize file change
    switch ext {
    case ".go":
        w.handleGoFileChange(filePath, event)
    case ".ts", ".tsx", ".js", ".jsx":
        w.handleFrontendFileChange(filePath, event)
    default:
        w.handleOtherFileChange(filePath, event)
    }
}

// handleGoFileChange processes Go file changes
func (w *EnhancedWatcher) handleGoFileChange(filePath string, event fsnotify.Event) {
    // Track modification time
    w.mu.Lock()
    w.goFiles[filePath] = time.Now()
    w.mu.Unlock()
    
    // Notify plugins
    if w.pluginManager != nil {
        ctx := &plugins.FileChangeContext{
            Context:     context.Background(),
            FilePath:    filePath,
            EventType:   eventTypeFromOp(event.Op),
            ProjectRoot: w.server.options.ProjectRoot,
            Logger:      w.logger,
        }
        
        // Execute file change hooks
        if err := w.pluginManager.ExecuteHook("OnFileChanged", ctx); err != nil {
            w.logger.Error("Plugin file change hook failed", "error", err)
        }
    }
    
    // Trigger rebuild
    w.server.triggerRebuild("go", filePath)
}

// handleFrontendFileChange processes frontend file changes
func (w *EnhancedWatcher) handleFrontendFileChange(filePath string, event fsnotify.Event) {
    // Check if it's a generated file
    if isGeneratedFile(filePath) {
        // Send HMR update
        w.server.sendHMRUpdate("generated", filePath)
        return
    }
    
    // Regular frontend file - standard HMR
    w.server.sendHMRUpdate("frontend", filePath)
}
```

### 2. HMR Protocol Implementation (`hmr.go`)
```go
package dev

import (
    "encoding/json"
    "fmt"
    "path/filepath"
    "strings"
)

// HMRMessage represents an HMR protocol message
type HMRMessage struct {
    Type      string                 `json:"type"`
    Timestamp int64                  `json:"timestamp"`
    Data      map[string]interface{} `json:"data"`
}

// HMR message types
const (
    HMRTypeReload      = "reload"
    HMRTypeUpdate      = "update"
    HMRTypeGenerated   = "generated"
    HMRTypeError       = "error"
    HMRTypeConnected   = "connected"
)

// sendHMRUpdate sends HMR update to connected clients
func (s *DevServer) sendHMRUpdate(updateType, filePath string) {
    relPath, _ := filepath.Rel(s.options.ProjectRoot, filePath)
    
    msg := HMRMessage{
        Type:      HMRTypeUpdate,
        Timestamp: time.Now().Unix(),
        Data: map[string]interface{}{
            "updateType": updateType,
            "file":       relPath,
            "preserve":   shouldPreserveState(filePath),
        },
    }
    
    // Special handling for generated files
    if updateType == "generated" {
        msg.Data["generated"] = true
        msg.Data["hooks"] = extractHooksFromFile(filePath)
    }
    
    s.broadcastHMR(msg)
}

// broadcastHMR sends message to all connected clients
func (s *DevServer) broadcastHMR(msg HMRMessage) {
    data, err := json.Marshal(msg)
    if err != nil {
        s.logger.Error("Failed to marshal HMR message", "error", err)
        return
    }
    
    s.wsHub.broadcast <- data
}

// shouldPreserveState determines if React state should be preserved
func shouldPreserveState(filePath string) bool {
    // Preserve state for:
    // - Style files
    // - Generated hook files
    // - Type definition files
    
    ext := filepath.Ext(filePath)
    base := filepath.Base(filePath)
    
    switch ext {
    case ".css", ".scss", ".less":
        return true
    case ".d.ts":
        return true
    }
    
    // Generated files preserve state
    if strings.Contains(filePath, "/generated/") {
        return true
    }
    
    // Hook files preserve state
    if strings.HasSuffix(base, "Hooks.ts") || strings.HasSuffix(base, "Hooks.tsx") {
        return true
    }
    
    return false
}

// extractHooksFromFile extracts hook names from generated file
func extractHooksFromFile(filePath string) []string {
    // Simple extraction - in real implementation, parse the file
    content, err := os.ReadFile(filePath)
    if err != nil {
        return nil
    }
    
    var hooks []string
    lines := strings.Split(string(content), "\n")
    
    for _, line := range lines {
        if strings.HasPrefix(strings.TrimSpace(line), "export function use") {
            // Extract hook name
            parts := strings.Fields(line)
            if len(parts) >= 3 {
                hookName := strings.TrimSuffix(parts[2], "(")
                hooks = append(hooks, hookName)
            }
        }
    }
    
    return hooks
}
```

### 3. WRF File Watching (`wrf/watcher.go`)
```go
package wrf

import (
    "path/filepath"
    "strings"
    
    "github.com/wailsapp/wails/v2/pkg/plugins"
)

// GetWatchPaths returns paths to watch for WRF plugin
func (p *WRFPlugin) GetWatchPaths(projectRoot string) ([]string, error) {
    var paths []string
    
    // Watch configured source paths
    for _, sourcePath := range p.config.Go.SourcePaths {
        absPath := filepath.Join(projectRoot, sourcePath)
        
        // Handle glob patterns
        if strings.Contains(sourcePath, "*") {
            matches, err := filepath.Glob(absPath)
            if err != nil {
                return nil, err
            }
            paths = append(paths, matches...)
        } else {
            paths = append(paths, absPath)
        }
    }
    
    // Watch WRF configuration
    paths = append(paths, filepath.Join(projectRoot, "wails.json"))
    
    return paths, nil
}

// OnFileChanged handles file changes for WRF plugin
func (p *WRFPlugin) OnFileChanged(ctx *plugins.FileChangeContext) error {
    // Only process Go files
    if !strings.HasSuffix(ctx.FilePath, ".go") {
        return nil
    }
    
    // Check if file is in watched paths
    if !p.isWatchedFile(ctx.FilePath) {
        return nil
    }
    
    p.logger.Info("WRF detected Go file change", "file", ctx.FilePath)
    
    // Mark for regeneration
    p.markForRegeneration(ctx.FilePath)
    
    // If in watch mode, trigger regeneration
    if p.config.Go.WatchMode {
        // Debounced regeneration
        p.scheduleRegeneration(ctx)
    }
    
    return nil
}

// scheduleRegeneration schedules code regeneration with debouncing
func (p *WRFPlugin) scheduleRegeneration(ctx *plugins.FileChangeContext) {
    p.regenerateDebouncer.Debounce("regenerate", func() {
        p.logger.Info("WRF triggering regeneration")
        
        genCtx := &plugins.GenerationContext{
            Context:       ctx.Context,
            ProjectRoot:   ctx.ProjectRoot,
            ProjectConfig: p.extractProjectConfig(),
            Logger:        ctx.Logger,
        }
        
        files, err := p.GenerateCode(genCtx)
        if err != nil {
            p.logger.Error("Regeneration failed", "error", err)
            
            // Send error to frontend
            p.sendHMRError(err)
            return
        }
        
        // Write generated files
        for _, file := range files {
            if err := p.writeGeneratedFile(file); err != nil {
                p.logger.Error("Failed to write file", "path", file.Path, "error", err)
            }
        }
        
        // Notify frontend of generated files
        p.sendGeneratedFilesUpdate(files)
    })
}
```

### 4. WRF HMR Integration (`wrf/hmr.go`)
```go
package wrf

import (
    "fmt"
    "path/filepath"
    "strings"
)

// sendGeneratedFilesUpdate notifies frontend of regenerated files
func (p *WRFPlugin) sendGeneratedFilesUpdate(files []*plugins.GeneratedFile) {
    var hooks []string
    var types []string
    var schemas []string
    
    // Categorize generated files
    for _, file := range files {
        relPath, _ := filepath.Rel(p.projectRoot, file.Path)
        
        if strings.Contains(relPath, "/hooks/") {
            hooks = append(hooks, relPath)
        } else if strings.Contains(relPath, "/types/") {
            types = append(types, relPath)
        } else if strings.Contains(relPath, "/schemas/") {
            schemas = append(schemas, relPath)
        }
    }
    
    // Send specialized HMR message
    msg := map[string]interface{}{
        "type": "wrf-regenerated",
        "data": map[string]interface{}{
            "hooks":   hooks,
            "types":   types,
            "schemas": schemas,
            "preserveState": true,
            "timestamp": time.Now().Unix(),
        },
    }
    
    p.devServer.BroadcastJSON(msg)
}

// sendHMRError sends error message to frontend
func (p *WRFPlugin) sendHMRError(err error) {
    msg := map[string]interface{}{
        "type": "wrf-error",
        "data": map[string]interface{}{
            "error":   err.Error(),
            "plugin":  "wrf",
            "action":  "regeneration",
            "timestamp": time.Now().Unix(),
        },
    }
    
    p.devServer.BroadcastJSON(msg)
}

// Frontend HMR client code (injected)
const hmrClientCode = `
// WRF HMR Client
(function() {
  const ws = new WebSocket('ws://localhost:34115/ws');
  
  ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    
    switch (msg.type) {
      case 'wrf-regenerated':
        handleWRFRegeneration(msg.data);
        break;
      case 'wrf-error':
        handleWRFError(msg.data);
        break;
      default:
        // Standard HMR handling
        break;
    }
  };
  
  function handleWRFRegeneration(data) {
    console.log('[WRF] Code regenerated', data);
    
    // Preserve React Query cache
    if (window.__REACT_QUERY_STATE__) {
      const cache = window.__REACT_QUERY_STATE__.getCache();
      
      // Update only affected queries
      data.hooks.forEach(hookFile => {
        // Invalidate queries related to this hook
        const hookName = extractHookName(hookFile);
        cache.invalidateQueries({ 
          predicate: (query) => query.queryKey[0] === hookName 
        });
      });
    }
    
    // Hot reload the modules
    if (module.hot) {
      data.hooks.forEach(file => {
        module.hot.accept('./' + file);
      });
      
      data.types.forEach(file => {
        module.hot.accept('./' + file);
      });
    }
  }
  
  function handleWRFError(data) {
    console.error('[WRF] Generation error:', data.error);
    
    // Show error overlay
    if (window.__WRF_ERROR_OVERLAY__) {
      window.__WRF_ERROR_OVERLAY__.showError({
        type: 'Generation Error',
        message: data.error,
        plugin: data.plugin,
        timestamp: new Date(data.timestamp * 1000)
      });
    }
  }
})();
`
```

### 5. Debouncer Utility
```go
// Debouncer prevents rapid repeated calls
type Debouncer struct {
    delay    time.Duration
    timers   map[string]*time.Timer
    mu       sync.Mutex
}

func NewDebouncer(delay time.Duration) *Debouncer {
    return &Debouncer{
        delay:  delay,
        timers: make(map[string]*time.Timer),
    }
}

func (d *Debouncer) Debounce(key string, fn func()) {
    d.mu.Lock()
    defer d.mu.Unlock()
    
    // Cancel existing timer
    if timer, exists := d.timers[key]; exists {
        timer.Stop()
    }
    
    // Create new timer
    d.timers[key] = time.AfterFunc(d.delay, func() {
        d.mu.Lock()
        delete(d.timers, key)
        d.mu.Unlock()
        
        fn()
    })
}
```

## Testing Strategy
- Test file watching with various file types
- Test debouncing with rapid changes
- Test HMR message protocol
- Test React state preservation
- Test error handling and recovery
- Performance test with many files

## Performance Considerations
- Debounce file events (500ms default)
- Batch regeneration for multiple files
- Efficient file filtering
- Minimal WebSocket traffic
- Smart cache invalidation

## Next Steps
With HMR complete, implement testing infrastructure for the plugin system.