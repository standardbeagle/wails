package dev

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wailsapp/wails/v2/pkg/clilogger"
	"github.com/wailsapp/wails/v2/pkg/plugins"
)

// mockDevServer implements basic dev server for testing
type mockDevServer struct {
	projectRoot     string
	triggeredBuilds []string
	hmrUpdates      []string
}

func (m *mockDevServer) triggerRebuild(buildType, filePath string) {
	m.triggeredBuilds = append(m.triggeredBuilds, buildType+":"+filePath)
}

func (m *mockDevServer) sendHMRUpdate(updateType, filePath string) {
	m.hmrUpdates = append(m.hmrUpdates, updateType+":"+filePath)
}

// mockPlugin implements FileWatcher for testing
type mockPlugin struct {
	name       string
	watchPaths []string
	changes    []string
}

func (m *mockPlugin) Name() string                                              { return m.name }
func (m *mockPlugin) Version() string                                           { return "1.0.0" }
func (m *mockPlugin) Description() string                                       { return "Mock plugin" }
func (m *mockPlugin) Author() string                                            { return "Test" }
func (m *mockPlugin) Initialize(ctx context.Context, config *plugins.PluginConfig) error { return nil }
func (m *mockPlugin) Shutdown(ctx context.Context) error                        { return nil }
func (m *mockPlugin) Health() error                                             { return nil }

func (m *mockPlugin) GetWatchPaths(projectRoot string) ([]string, error) {
	return m.watchPaths, nil
}

func (m *mockPlugin) OnFileChanged(ctx *plugins.FileChangeContext) error {
	m.changes = append(m.changes, ctx.FilePath)
	return nil
}

// mockPluginManager implements basic plugin manager for testing
type mockPluginManager struct {
	plugins []plugins.Plugin
}

func (m *mockPluginManager) GetPlugins() []plugins.Plugin {
	return m.plugins
}

func (m *mockPluginManager) ExecuteHook(hookName string, ctx interface{}) error {
	for _, plugin := range m.plugins {
		if fw, ok := plugin.(plugins.FileWatcher); ok {
			if hookName == "OnFileChanged" {
				if changeCtx, ok := ctx.(*plugins.FileChangeContext); ok {
					fw.OnFileChanged(changeCtx)
				}
			}
		}
	}
	return nil
}

func TestEnhancedWatcher_FileFiltering(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     bool
	}{
		{"ignore hidden files", ".hidden", true},
		{"ignore temp files", "file.tmp", true},
		{"ignore backup files", "file~", true},
		{"ignore node_modules", "node_modules/package", true},
		{"ignore build dir", "build/bin/app", true},
		{"allow go files", "main.go", false},
		{"allow ts files", "app.ts", false},
		{"allow regular files", "README.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldIgnore(tt.filePath)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestEnhancedWatcher_EventTypeConversion(t *testing.T) {
	tests := []struct {
		name string
		op   fsnotify.Op
		want string
	}{
		{"create", fsnotify.Create, "created"},
		{"write", fsnotify.Write, "modified"},
		{"remove", fsnotify.Remove, "deleted"},
		{"rename", fsnotify.Rename, "renamed"},
		{"chmod", fsnotify.Chmod, "modified"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eventTypeFromOp(tt.op)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestEnhancedWatcher_GeneratedFileDetection(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     bool
	}{
		{"generated directory", "src/generated/hooks.ts", true},
		{"hooks directory", "src/hooks/useUser.ts", true},
		{"types file", "src/types.ts", true},
		{"schemas file", "src/schemas.ts", true},
		{"regular file", "src/components/App.tsx", false},
		{"regular ts file", "src/utils.ts", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isGeneratedFile(tt.filePath)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestDebouncer(t *testing.T) {
	debouncer := NewDebouncer(100 * time.Millisecond)
	
	var callCount int
	key := "test"
	
	// Multiple rapid calls should only execute once
	for i := 0; i < 5; i++ {
		debouncer.Debounce(key, func() {
			callCount++
		})
		time.Sleep(50 * time.Millisecond) // Less than debounce delay
	}
	
	// Wait for debounce to complete
	time.Sleep(200 * time.Millisecond)
	
	assert.Equal(t, 1, callCount, "Debouncer should only execute once")
	
	// Test HasPending
	debouncer.Debounce("test2", func() {})
	assert.True(t, debouncer.HasPending())
	
	// Test Clear
	debouncer.Clear()
	assert.False(t, debouncer.HasPending())
}

func TestEnhancedWatcher_Integration(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	
	// Create mock server
	server := &mockDevServer{
		projectRoot: tmpDir,
	}
	
	// Create mock plugin
	plugin := &mockPlugin{
		name:       "test-plugin",
		watchPaths: []string{filepath.Join(tmpDir, "app")},
	}
	
	// Create mock plugin manager
	manager := &mockPluginManager{
		plugins: []plugins.Plugin{plugin},
	}
	
	// Create logger
	logger := clilogger.New(nil)
	logger.Mute(true)
	
	// Create enhanced watcher
	watcher, err := NewEnhancedWatcher(server, manager, logger)
	require.NoError(t, err)
	require.NotNil(t, watcher)
	
	// Test that plugin is registered
	assert.Equal(t, "test-plugin", plugin.name)
	assert.Equal(t, []string{filepath.Join(tmpDir, "app")}, plugin.watchPaths)
	
	// Test debouncer creation
	assert.NotNil(t, watcher.debouncer)
	assert.Equal(t, 500*time.Millisecond, watcher.debouncer.delay)
	
	// Close watcher
	err = watcher.Close()
	assert.NoError(t, err)
}

func TestEnhancedWatcher_GoFileHandling(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	
	// Create mock server
	server := &mockDevServer{
		projectRoot: tmpDir,
	}
	
	// Create mock plugin
	plugin := &mockPlugin{
		name:       "test-plugin",
		watchPaths: []string{tmpDir},
	}
	
	// Create mock plugin manager
	manager := &mockPluginManager{
		plugins: []plugins.Plugin{plugin},
	}
	
	// Create logger
	logger := clilogger.New(nil)
	logger.Mute(true)
	
	// Create enhanced watcher
	watcher, err := NewEnhancedWatcher(server, manager, logger)
	require.NoError(t, err)
	
	// Create test file
	testFile := filepath.Join(tmpDir, "test.go")
	err = os.WriteFile(testFile, []byte("package main"), 0644)
	require.NoError(t, err)
	
	// Simulate file change event
	event := fsnotify.Event{
		Name: testFile,
		Op:   fsnotify.Write,
	}
	
	// Process the file change
	watcher.processFileChange(event)
	
	// Allow time for processing
	time.Sleep(100 * time.Millisecond)
	
	// Check that rebuild was triggered
	assert.Contains(t, server.triggeredBuilds, "go:"+testFile)
	
	// Check that plugin was notified
	assert.Contains(t, plugin.changes, testFile)
	
	// Close watcher
	err = watcher.Close()
	assert.NoError(t, err)
}