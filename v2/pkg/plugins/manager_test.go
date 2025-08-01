package plugins

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestNewManager tests the creation of a new plugin manager.
func TestNewManager(t *testing.T) {
	// Test with nil config
	m1 := NewManager(nil)
	if m1 == nil {
		t.Fatal("NewManager returned nil with nil config")
	}
	if m1.config.MaxInitTimeout != 30*time.Second {
		t.Errorf("Expected default init timeout of 30s, got %v", m1.config.MaxInitTimeout)
	}

	// Test with custom config
	config := &ManagerConfig{
		MaxInitTimeout:      60 * time.Second,
		MaxShutdownTimeout:  20 * time.Second,
		EnableParallelHooks: false,
	}
	m2 := NewManager(config)
	if m2.config.MaxInitTimeout != 60*time.Second {
		t.Errorf("Expected init timeout of 60s, got %v", m2.config.MaxInitTimeout)
	}
	if m2.config.EnableParallelHooks != false {
		t.Error("Expected parallel hooks to be disabled")
	}
}

// TestPluginRegistration tests plugin registration functionality.
func TestPluginRegistration(t *testing.T) {
	m := NewManager(nil)

	// Test registering a valid plugin
	plugin1 := NewMockPlugin("test-plugin-1")
	err := m.RegisterPlugin(plugin1)
	if err != nil {
		t.Errorf("Failed to register valid plugin: %v", err)
	}

	// Test registering nil plugin
	err = m.RegisterPlugin(nil)
	if err == nil {
		t.Error("Expected error when registering nil plugin")
	}

	// Test registering plugin with empty name
	badPlugin := &MockPlugin{version: "1.0.0"}
	err = m.RegisterPlugin(badPlugin)
	if err == nil {
		t.Error("Expected error when registering plugin with empty name")
	}

	// Test registering duplicate plugin
	plugin2 := NewMockPlugin("test-plugin-1")
	err = m.RegisterPlugin(plugin2)
	if err == nil {
		t.Error("Expected error when registering duplicate plugin")
	}

	// Test GetPlugin
	retrieved, err := m.GetPlugin("test-plugin-1")
	if err != nil {
		t.Errorf("Failed to get registered plugin: %v", err)
	}
	if retrieved.Name() != "test-plugin-1" {
		t.Errorf("Retrieved wrong plugin: %s", retrieved.Name())
	}

	// Test GetPlugin with non-existent plugin
	_, err = m.GetPlugin("non-existent")
	if err == nil {
		t.Error("Expected error when getting non-existent plugin")
	}
}

// TestPluginLifecycle tests plugin initialization and shutdown.
func TestPluginLifecycle(t *testing.T) {
	m := NewManager(&ManagerConfig{
		ProjectRoot: "/test/project",
		ProjectName: "test-project",
	})

	plugin := NewMockPlugin("lifecycle-test")
	err := m.RegisterPlugin(plugin)
	if err != nil {
		t.Fatalf("Failed to register plugin: %v", err)
	}

	// Test initialization
	ctx := context.Background()
	err = m.Initialize(ctx)
	if err != nil {
		t.Errorf("Failed to initialize manager: %v", err)
	}

	if !plugin.initialized {
		t.Error("Plugin was not initialized")
	}

	// Test state after initialization
	state, err := m.GetPluginState("lifecycle-test")
	if err != nil {
		t.Errorf("Failed to get plugin state: %v", err)
	}
	if state != StateReady {
		t.Errorf("Expected plugin state to be Ready, got %v", state)
	}

	// Test double initialization (should be no-op)
	err = m.Initialize(ctx)
	if err != nil {
		t.Errorf("Second initialization failed: %v", err)
	}

	// Test shutdown
	err = m.Shutdown(ctx)
	if err != nil {
		t.Errorf("Failed to shutdown manager: %v", err)
	}

	if !plugin.shutdown {
		t.Error("Plugin was not shut down")
	}

	// Test state after shutdown
	state, err = m.GetPluginState("lifecycle-test")
	if err != nil {
		t.Errorf("Failed to get plugin state: %v", err)
	}
	if state != StateStopped {
		t.Errorf("Expected plugin state to be Stopped, got %v", state)
	}
}

// TestGetPluginsByType tests retrieving plugins by interface type.
func TestGetPluginsByType(t *testing.T) {
	m := NewManager(nil)

	// Register different types of plugins
	basicPlugin := NewMockPlugin("basic")
	buildPlugin := NewMockBuildHookPlugin("build")
	codeGenPlugin := NewMockCodeGeneratorPlugin("codegen")

	m.RegisterPlugin(basicPlugin)
	m.RegisterPlugin(buildPlugin)
	m.RegisterPlugin(codeGenPlugin)

	// Initialize all plugins
	ctx := context.Background()
	m.Initialize(ctx)

	// Test getting BuildHook plugins
	buildHooks := m.GetPluginsByType((*BuildHook)(nil))
	if len(buildHooks) != 1 {
		t.Errorf("Expected 1 BuildHook plugin, got %d", len(buildHooks))
	}

	// Test getting CodeGenerator plugins
	codeGens := m.GetPluginsByType((*CodeGenerator)(nil))
	if len(codeGens) != 1 {
		t.Errorf("Expected 1 CodeGenerator plugin, got %d", len(codeGens))
	}

	// Test getting non-existent interface type
	fileWatchers := m.GetPluginsByType((*FileWatcher)(nil))
	if len(fileWatchers) != 0 {
		t.Errorf("Expected 0 FileWatcher plugins, got %d", len(fileWatchers))
	}
}

// TestExecuteHook tests hook execution functionality.
func TestExecuteHook(t *testing.T) {
	m := NewManager(&ManagerConfig{
		EnableParallelHooks: false, // Test sequential execution first
	})

	// Register BuildHook plugins
	plugin1 := NewMockBuildHookPlugin("build-1")
	plugin2 := NewMockBuildHookPlugin("build-2")

	m.RegisterPlugin(plugin1)
	m.RegisterPlugin(plugin2)
	m.Initialize(context.Background())

	// Test PreBuild hook
	buildCtx := &BuildContext{
		Context:     context.Background(),
		ProjectRoot: "/test",
		BuildMode:   "production",
	}

	err := m.ExecuteHook("PreBuild", buildCtx)
	if err != nil {
		t.Errorf("Failed to execute PreBuild hook: %v", err)
	}

	if !plugin1.preBuildCalled || !plugin2.preBuildCalled {
		t.Error("Not all PreBuild hooks were called")
	}

	// Test PostBuild hook
	err = m.ExecuteHook("PostBuild", buildCtx)
	if err != nil {
		t.Errorf("Failed to execute PostBuild hook: %v", err)
	}

	if !plugin1.postBuildCalled || !plugin2.postBuildCalled {
		t.Error("Not all PostBuild hooks were called")
	}

	// Test unknown hook
	err = m.ExecuteHook("UnknownHook", buildCtx)
	if err == nil {
		t.Error("Expected error for unknown hook")
	}
}

// TestParallelHookExecution tests parallel hook execution.
func TestParallelHookExecution(t *testing.T) {
	m := NewManager(&ManagerConfig{
		EnableParallelHooks: true,
	})

	// Create plugins that track execution order
	var executionOrder []string
	var orderMu sync.Mutex

	plugin1 := &SlowMockPlugin{
		MockPlugin: MockPlugin{
			name:    "slow-1",
			version: "1.0.0",
		},
		delay: 50 * time.Millisecond,
		onExecute: func(name string) {
			orderMu.Lock()
			executionOrder = append(executionOrder, name)
			orderMu.Unlock()
		},
	}

	plugin2 := &SlowMockPlugin{
		MockPlugin: MockPlugin{
			name:    "slow-2",
			version: "1.0.0",
		},
		delay: 10 * time.Millisecond,
		onExecute: func(name string) {
			orderMu.Lock()
			executionOrder = append(executionOrder, name)
			orderMu.Unlock()
		},
	}

	m.RegisterPlugin(plugin1)
	m.RegisterPlugin(plugin2)
	m.Initialize(context.Background())

	// Execute a parallel-safe hook
	ctx := &FileChangeContext{
		Context:     context.Background(),
		FilePath:    "/test/file.go",
		ChangeType:  "modified",
	}

	err := m.ExecuteHook("OnFileChanged", ctx)
	if err != nil {
		t.Errorf("Failed to execute parallel hook: %v", err)
	}

	// With parallel execution, plugin2 (faster) should complete first
	if len(executionOrder) != 2 {
		t.Errorf("Expected 2 executions, got %d", len(executionOrder))
	}
}

// TestConcurrentAccess tests thread-safe concurrent access to the manager.
func TestConcurrentAccess(t *testing.T) {
	m := NewManager(nil)

	// Use atomic counter for race detection
	var registerCount int32
	var getCount int32

	// Run concurrent operations
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(3)

		// Concurrent registration
		go func(id int) {
			defer wg.Done()
			plugin := NewMockPlugin(fmt.Sprintf("concurrent-%d", id))
			m.RegisterPlugin(plugin)
			atomic.AddInt32(&registerCount, 1)
		}(i)

		// Concurrent retrieval
		go func(id int) {
			defer wg.Done()
			time.Sleep(10 * time.Millisecond) // Give registration a chance
			_, err := m.GetPlugin(fmt.Sprintf("concurrent-%d", id))
			if err == nil {
				atomic.AddInt32(&getCount, 1)
			}
		}(i)

		// Concurrent type queries
		go func() {
			defer wg.Done()
			m.GetPluginsByType((*Plugin)(nil))
		}()
	}

	wg.Wait()

	// Verify results
	if registerCount != 10 {
		t.Errorf("Expected 10 registrations, got %d", registerCount)
	}
}

// TestPluginPanicRecovery tests that the manager recovers from plugin panics.
func TestPluginPanicRecovery(t *testing.T) {
	m := NewManager(nil)

	// Create a plugin that panics
	panicPlugin := &PanicPlugin{
		MockPlugin: MockPlugin{
			name:    "panic-plugin",
			version: "1.0.0",
		},
	}

	m.RegisterPlugin(panicPlugin)
	m.Initialize(context.Background())

	// Execute hook that will panic
	ctx := &BuildContext{
		Context: context.Background(),
	}

	// This should not panic the manager
	err := m.ExecuteHook("PreBuild", ctx)
	if err == nil {
		t.Error("Expected error from panicking plugin")
	}

	// Manager should still be functional
	_, err = m.GetPlugin("panic-plugin")
	if err != nil {
		t.Error("Manager is not functional after plugin panic")
	}
}

// TestHealthCheck tests the health check functionality.
func TestHealthCheck(t *testing.T) {
	m := NewManager(nil)

	// Register healthy and unhealthy plugins
	healthyPlugin := NewMockPlugin("healthy")
	unhealthyPlugin := &MockPlugin{
		name:        "unhealthy",
		version:     "1.0.0",
		healthError: errors.New("plugin is unhealthy"),
	}

	m.RegisterPlugin(healthyPlugin)
	m.RegisterPlugin(unhealthyPlugin)
	m.Initialize(context.Background())

	// Check health
	err := m.Health()
	if err == nil {
		t.Error("Expected health check to fail with unhealthy plugin")
	}
	if err.Error() == "" {
		t.Error("Health check error should contain details")
	}
}

// TestPluginStateTransitions tests plugin state management.
func TestPluginStateTransitions(t *testing.T) {
	m := NewManager(nil)
	
	plugin := NewMockPlugin("state-test")
	m.RegisterPlugin(plugin)

	// Check initial state
	state, err := m.GetPluginState("state-test")
	if err != nil {
		t.Errorf("Failed to get initial state: %v", err)
	}
	if state != StateUninitialized {
		t.Errorf("Expected initial state Uninitialized, got %v", state)
	}

	// Test invalid state request
	_, err = m.GetPluginState("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent plugin state")
	}
}

// Helper types for testing

// SlowMockPlugin simulates a slow plugin for testing parallel execution.
type SlowMockPlugin struct {
	MockPlugin
	delay     time.Duration
	onExecute func(string)
}

func (p *SlowMockPlugin) Initialize(ctx context.Context, config *PluginConfig) error {
	p.initialized = true
	return nil
}

func (p *SlowMockPlugin) GetWatchPaths(projectRoot string) ([]string, error) {
	return []string{projectRoot}, nil
}

func (p *SlowMockPlugin) OnFileChanged(ctx *FileChangeContext) error {
	time.Sleep(p.delay)
	if p.onExecute != nil {
		p.onExecute(p.name)
	}
	return nil
}

// Implement FileWatcher interface
var _ FileWatcher = (*SlowMockPlugin)(nil)

// PanicPlugin is a plugin that panics during execution.
type PanicPlugin struct {
	MockPlugin
}

func (p *PanicPlugin) Initialize(ctx context.Context, config *PluginConfig) error {
	p.initialized = true
	return nil
}

func (p *PanicPlugin) PreBuild(ctx *BuildContext) error {
	panic("intentional panic for testing")
}

func (p *PanicPlugin) PostBuild(ctx *BuildContext) error {
	return nil
}

// Implement BuildHook interface
var _ BuildHook = (*PanicPlugin)(nil)

// BenchmarkGetPluginsByType benchmarks the performance of GetPluginsByType.
func BenchmarkGetPluginsByType(b *testing.B) {
	m := NewManager(nil)

	// Register 100 plugins of different types
	for i := 0; i < 50; i++ {
		m.RegisterPlugin(NewMockBuildHookPlugin(fmt.Sprintf("build-%d", i)))
		m.RegisterPlugin(NewMockCodeGeneratorPlugin(fmt.Sprintf("codegen-%d", i)))
	}

	m.Initialize(context.Background())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.GetPluginsByType((*BuildHook)(nil))
	}
}

// BenchmarkConcurrentPluginAccess benchmarks concurrent plugin access.
func BenchmarkConcurrentPluginAccess(b *testing.B) {
	m := NewManager(nil)

	// Register plugins
	for i := 0; i < 10; i++ {
		m.RegisterPlugin(NewMockPlugin(fmt.Sprintf("plugin-%d", i)))
	}

	m.Initialize(context.Background())

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			name := fmt.Sprintf("plugin-%d", i%10)
			_, _ = m.GetPlugin(name)
			i++
		}
	})
}