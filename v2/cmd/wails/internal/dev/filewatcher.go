package dev

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/wailsapp/wails/v2/pkg/plugins"
	"github.com/wailsapp/wails/v2/pkg/clilogger"
)

// EnhancedWatcher wraps fsnotify with plugin support
type EnhancedWatcher struct {
	watcher       *fsnotify.Watcher
	pluginManager *plugins.Manager
	server        *DevServer
	logger        *clilogger.CLILogger

	// Debouncing
	debouncer *Debouncer

	// File categorization
	goFiles     map[string]time.Time
	pluginFiles map[string]string // file -> plugin name

	mu sync.RWMutex
}

// NewEnhancedWatcher creates enhanced file watcher
func NewEnhancedWatcher(server *DevServer, pluginManager *plugins.Manager, logger *clilogger.CLILogger) (*EnhancedWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &EnhancedWatcher{
		watcher:       watcher,
		pluginManager: pluginManager,
		server:        server,
		logger:        logger,
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

// addStandardPaths adds standard Wails paths to watch
func (w *EnhancedWatcher) addStandardPaths() error {
	// Standard Go source paths
	paths := []string{
		".",
		"app",
		"build",
	}

	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			continue
		}

		if err := w.watcher.Add(absPath); err != nil {
			w.logger.Debug("Failed to watch standard path", "path", path, "error", err)
		} else {
			w.logger.Debug("Watching standard path", "path", path)
		}
	}

	return nil
}

// addPluginPaths adds paths from FileWatcher plugins
func (w *EnhancedWatcher) addPluginPaths() error {
	// Get all plugins that implement FileWatcher interface
	allPlugins := w.pluginManager.GetPlugins()

	for _, plugin := range allPlugins {
		if fw, ok := plugin.(plugins.FileWatcher); ok {
			paths, err := fw.GetWatchPaths(w.server.projectRoot)
			if err != nil {
				w.logger.Warn("Failed to get plugin paths", "plugin", plugin.Name(), "error", err)
				continue
			}

			for _, path := range paths {
				absPath, _ := filepath.Abs(path)
				if err := w.watcher.Add(absPath); err != nil {
					w.logger.Warn("Failed to watch plugin path", "path", path, "error", err)
				} else {
					// Track which plugin registered this path
					w.mu.Lock()
					w.pluginFiles[absPath] = plugin.Name()
					w.mu.Unlock()

					w.logger.Debug("Watching for plugin", "path", path, "plugin", plugin.Name())
				}
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
			ChangeType:  eventTypeFromOp(event.Op),
			ProjectRoot: w.server.projectRoot,
			Timestamp:   time.Now(),
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

// handleOtherFileChange processes other file changes
func (w *EnhancedWatcher) handleOtherFileChange(filePath string, event fsnotify.Event) {
	// Handle wails.json changes
	if strings.HasSuffix(filePath, "wails.json") {
		w.logger.Info("Configuration file changed, triggering full rebuild")
		w.server.triggerRebuild("config", filePath)
	}
}

// shouldIgnore determines if a file should be ignored
func shouldIgnore(filePath string) bool {
	base := filepath.Base(filePath)

	// Ignore hidden files
	if strings.HasPrefix(base, ".") {
		return true
	}

	// Ignore temporary files
	if strings.HasSuffix(base, "~") || strings.HasSuffix(base, ".tmp") {
		return true
	}

	// Ignore build artifacts
	ignored := []string{
		"build/bin/",
		"dist/",
		"node_modules/",
		".git/",
		"__pycache__/",
	}

	for _, ignore := range ignored {
		if strings.Contains(filePath, ignore) {
			return true
		}
	}

	return false
}

// eventTypeFromOp converts fsnotify.Op to string
func eventTypeFromOp(op fsnotify.Op) string {
	switch {
	case op&fsnotify.Create == fsnotify.Create:
		return "created"
	case op&fsnotify.Write == fsnotify.Write:
		return "modified"
	case op&fsnotify.Remove == fsnotify.Remove:
		return "deleted"
	case op&fsnotify.Rename == fsnotify.Rename:
		return "renamed"
	case op&fsnotify.Chmod == fsnotify.Chmod:
		return "modified"
	default:
		return "modified"
	}
}

// isGeneratedFile checks if a file is generated by plugins
func isGeneratedFile(filePath string) bool {
	return strings.Contains(filePath, "/generated/") ||
		strings.Contains(filePath, "/hooks/") ||
		strings.Contains(filePath, "/types.ts") ||
		strings.Contains(filePath, "/schemas.ts")
}

// Close stops the watcher
func (w *EnhancedWatcher) Close() error {
	if w.watcher != nil {
		return w.watcher.Close()
	}
	return nil
}