# Task: Testing Infrastructure
**Generated from Master Planning**: 2024-01-15
**Context Package**: `/requests/wails-plugin-fork/context/`

## Task Sizing Assessment
**File Count**: 6-8 files
**Estimated Time**: 60 minutes
**Token Estimate**: 100k tokens
**Complexity Level**: 3 (Complex)
**Parallelization Benefit**: HIGH - Independent test suites
**Atomicity Assessment**: ✅ ATOMIC - Complete test framework
**Boundary Analysis**: ✅ CLEAR - Test files only

## Persona Assignment
**Persona**: QA Engineer
**Expertise Required**: Go testing, mocking, integration testing
**Worktree**: `~/work/wrf/wails-plugin-fork/`

## Context Summary
**Risk Level**: LOW - Tests don't affect production code
**Integration Points**: All plugin system components
**Architecture Pattern**: Table-driven tests, test helpers
**Similar Reference**: Existing Wails test patterns

### Task Scope Boundaries
**MODIFY Zone** (Direct Changes):
```yaml
primary_files:
  - /v2/pkg/plugins/manager_test.go         # Manager tests
  - /v2/pkg/plugins/plugin_test.go          # Interface tests
  - /v2/pkg/plugins/testutil/mocks.go       # Mock implementations
  - /v2/pkg/plugins/integration_test.go     # Integration tests
  - /v2/pkg/plugins/builtin/react/react_test.go    # React plugin tests
  - /v2/pkg/plugins/wrf/wrf_test.go         # WRF plugin tests
  - /v2/cmd/wails/internal/commands/plugin/plugin_test.go  # CLI tests
  - /v2/test/fixtures/                      # Test fixtures
```

**REVIEW Zone** (Check for Impact):
```yaml
check_patterns:
  - /v2/internal/test/                     # Test utilities
  - /v2/Makefile                           # Test targets
```

**IGNORE Zone** (Do Not Touch):
```yaml
ignore_completely:
  - /v2/build/                             # Build artifacts
  - /v2/internal/frontend/                 # Frontend code
```

## Task Requirements
**Objective**: Comprehensive test coverage for plugin system following Wails standards

**Success Criteria**:
- [ ] Unit tests for all plugin components
- [ ] Integration tests for plugin lifecycle
- [ ] Mock plugins for testing
- [ ] Test helpers and utilities
- [ ] Performance benchmarks
- [ ] Race condition tests
- [ ] >80% code coverage

**Validation Commands**:
```bash
# Run all plugin tests
go test ./v2/pkg/plugins/...

# Run with coverage
go test -cover ./v2/pkg/plugins/...

# Run with race detection
go test -race ./v2/pkg/plugins/...

# Run benchmarks
go test -bench=. ./v2/pkg/plugins/...

# Generate coverage report
go test -coverprofile=coverage.out ./v2/pkg/plugins/...
go tool cover -html=coverage.out
```

## Implementation Details

### 1. Test Utilities and Mocks (`testutil/mocks.go`)
```go
package testutil

import (
    "context"
    "sync"
    "github.com/stretchr/testify/mock"
    "github.com/wailsapp/wails/v2/pkg/plugins"
)

// MockPlugin provides a mock plugin for testing
type MockPlugin struct {
    mock.Mock
    mu sync.Mutex
    
    // Tracking
    InitCalled     bool
    ShutdownCalled bool
    HealthCalled   int
}

func NewMockPlugin(name, version string) *MockPlugin {
    m := &MockPlugin{}
    m.On("Name").Return(name)
    m.On("Version").Return(version)
    m.On("Description").Return("Mock plugin for testing")
    m.On("Author").Return("Test")
    return m
}

func (m *MockPlugin) Name() string {
    args := m.Called()
    return args.String(0)
}

func (m *MockPlugin) Version() string {
    args := m.Called()
    return args.String(0)
}

func (m *MockPlugin) Description() string {
    args := m.Called()
    return args.String(0)
}

func (m *MockPlugin) Author() string {
    args := m.Called()
    return args.String(0)
}

func (m *MockPlugin) Initialize(ctx context.Context, config *plugins.PluginConfig) error {
    m.mu.Lock()
    m.InitCalled = true
    m.mu.Unlock()
    
    args := m.Called(ctx, config)
    return args.Error(0)
}

func (m *MockPlugin) Shutdown(ctx context.Context) error {
    m.mu.Lock()
    m.ShutdownCalled = true
    m.mu.Unlock()
    
    args := m.Called(ctx)
    return args.Error(0)
}

func (m *MockPlugin) Health() error {
    m.mu.Lock()
    m.HealthCalled++
    m.mu.Unlock()
    
    args := m.Called()
    return args.Error(0)
}

// MockBuildHook provides a mock build hook plugin
type MockBuildHook struct {
    MockPlugin
    PreBuildCalled  bool
    PostBuildCalled bool
}

func NewMockBuildHook() *MockBuildHook {
    m := &MockBuildHook{
        MockPlugin: *NewMockPlugin("mock-build", "1.0.0"),
    }
    return m
}

func (m *MockBuildHook) PreBuild(ctx *plugins.BuildContext) error {
    m.mu.Lock()
    m.PreBuildCalled = true
    m.mu.Unlock()
    
    args := m.Called(ctx)
    return args.Error(0)
}

func (m *MockBuildHook) PostBuild(ctx *plugins.BuildContext) error {
    m.mu.Lock()
    m.PostBuildCalled = true
    m.mu.Unlock()
    
    args := m.Called(ctx)
    return args.Error(0)
}

// MockCodeGenerator provides a mock code generator
type MockCodeGenerator struct {
    MockPlugin
    GenerateCodeCalled bool
    GeneratedFiles     []*plugins.GeneratedFile
}

func NewMockCodeGenerator() *MockCodeGenerator {
    m := &MockCodeGenerator{
        MockPlugin: *NewMockPlugin("mock-codegen", "1.0.0"),
    }
    return m
}

func (m *MockCodeGenerator) GenerateCode(ctx *plugins.GenerationContext) ([]*plugins.GeneratedFile, error) {
    m.mu.Lock()
    m.GenerateCodeCalled = true
    m.mu.Unlock()
    
    args := m.Called(ctx)
    
    if files, ok := args.Get(0).([]*plugins.GeneratedFile); ok {
        return files, args.Error(1)
    }
    
    return m.GeneratedFiles, args.Error(1)
}

func (m *MockCodeGenerator) SupportedLanguages() []string {
    args := m.Called()
    return args.Get(0).([]string)
}

// Test helpers
type TestLogger struct {
    mu       sync.Mutex
    messages []LogMessage
}

type LogMessage struct {
    Level   string
    Message string
    Fields  map[string]interface{}
}

func NewTestLogger() *TestLogger {
    return &TestLogger{
        messages: make([]LogMessage, 0),
    }
}

func (l *TestLogger) Info(msg string, args ...interface{}) {
    l.log("INFO", msg, args...)
}

func (l *TestLogger) Error(msg string, args ...interface{}) {
    l.log("ERROR", msg, args...)
}

func (l *TestLogger) Debug(msg string, args ...interface{}) {
    l.log("DEBUG", msg, args...)
}

func (l *TestLogger) Warn(msg string, args ...interface{}) {
    l.log("WARN", msg, args...)
}

func (l *TestLogger) log(level, msg string, args ...interface{}) {
    l.mu.Lock()
    defer l.mu.Unlock()
    
    fields := make(map[string]interface{})
    for i := 0; i < len(args); i += 2 {
        if i+1 < len(args) {
            if key, ok := args[i].(string); ok {
                fields[key] = args[i+1]
            }
        }
    }
    
    l.messages = append(l.messages, LogMessage{
        Level:   level,
        Message: msg,
        Fields:  fields,
    })
}

func (l *TestLogger) HasMessage(level, msg string) bool {
    l.mu.Lock()
    defer l.mu.Unlock()
    
    for _, m := range l.messages {
        if m.Level == level && m.Message == msg {
            return true
        }
    }
    return false
}
```

### 2. Manager Tests (`manager_test.go`)
```go
package plugins

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/wailsapp/wails/v2/pkg/plugins/testutil"
)

func TestManager_Initialize(t *testing.T) {
    tests := []struct {
        name    string
        config  *ManagerConfig
        plugins []Plugin
        wantErr bool
    }{
        {
            name: "successful initialization",
            config: &ManagerConfig{
                Plugins: map[string]*PluginEntry{
                    "test-plugin": {
                        Enabled: true,
                        Config:  map[string]interface{}{"key": "value"},
                    },
                },
                LoadTimeout: 5 * time.Second,
            },
            plugins: []Plugin{
                testutil.NewMockPlugin("test-plugin", "1.0.0"),
            },
            wantErr: false,
        },
        {
            name: "plugin not found",
            config: &ManagerConfig{
                Plugins: map[string]*PluginEntry{
                    "missing-plugin": {Enabled: true},
                },
            },
            plugins: []Plugin{},
            wantErr: false, // Should log warning but not fail
        },
        {
            name: "plugin initialization error",
            config: &ManagerConfig{
                Plugins: map[string]*PluginEntry{
                    "error-plugin": {Enabled: true},
                },
            },
            plugins: []Plugin{
                func() Plugin {
                    m := testutil.NewMockPlugin("error-plugin", "1.0.0")
                    m.On("Initialize", mock.Anything, mock.Anything).Return(errors.New("init error"))
                    return m
                }(),
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            logger := testutil.NewTestLogger()
            manager := NewManager(tt.config, logger)
            
            // Register test plugins
            for _, p := range tt.plugins {
                err := manager.registerPlugin(p)
                require.NoError(t, err)
            }
            
            ctx := context.Background()
            wailsConfig := &options.App{Name: "test-app"}
            
            err := manager.Initialize(ctx, wailsConfig)
            
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                
                // Verify plugins were initialized
                for _, p := range tt.plugins {
                    if mp, ok := p.(*testutil.MockPlugin); ok {
                        if tt.config.Plugins[p.Name()].Enabled {
                            assert.True(t, mp.InitCalled)
                        }
                    }
                }
            }
        })
    }
}

func TestManager_ExecuteHook(t *testing.T) {
    logger := testutil.NewTestLogger()
    config := &ManagerConfig{}
    manager := NewManager(config, logger)
    
    // Create mock plugins
    buildHook := testutil.NewMockBuildHook()
    buildHook.On("PreBuild", mock.Anything).Return(nil)
    buildHook.On("PostBuild", mock.Anything).Return(nil)
    
    codeGen := testutil.NewMockCodeGenerator()
    codeGen.On("GenerateCode", mock.Anything).Return([]*GeneratedFile{
        {Path: "test.ts", Content: []byte("// test")},
    }, nil)
    
    // Register plugins
    require.NoError(t, manager.registerPlugin(buildHook))
    require.NoError(t, manager.registerPlugin(codeGen))
    
    // Initialize
    ctx := context.Background()
    wailsConfig := &options.App{Name: "test-app"}
    require.NoError(t, manager.Initialize(ctx, wailsConfig))
    
    // Test build hooks
    buildCtx := &BuildContext{
        Context:     ctx,
        ProjectRoot: "/test",
        BuildMode:   "dev",
    }
    
    err := manager.ExecuteHook("PreBuild", buildCtx)
    assert.NoError(t, err)
    assert.True(t, buildHook.PreBuildCalled)
    
    err = manager.ExecuteHook("PostBuild", buildCtx)
    assert.NoError(t, err)
    assert.True(t, buildHook.PostBuildCalled)
    
    // Test code generation
    genCtx := &GenerationContext{
        Context:     ctx,
        ProjectRoot: "/test",
    }
    
    err = manager.ExecuteHook("GenerateCode", genCtx)
    assert.NoError(t, err)
    assert.True(t, codeGen.GenerateCodeCalled)
}

func TestManager_Concurrency(t *testing.T) {
    logger := testutil.NewTestLogger()
    config := &ManagerConfig{}
    manager := NewManager(config, logger)
    
    // Create multiple plugins
    var plugins []Plugin
    for i := 0; i < 10; i++ {
        p := testutil.NewMockPlugin(fmt.Sprintf("plugin-%d", i), "1.0.0")
        p.On("Initialize", mock.Anything, mock.Anything).Return(nil).Maybe()
        p.On("Shutdown", mock.Anything).Return(nil).Maybe()
        p.On("Health").Return(nil).Maybe()
        plugins = append(plugins, p)
    }
    
    // Register plugins concurrently
    var wg sync.WaitGroup
    for _, p := range plugins {
        wg.Add(1)
        go func(plugin Plugin) {
            defer wg.Done()
            err := manager.registerPlugin(plugin)
            assert.NoError(t, err)
        }(p)
    }
    wg.Wait()
    
    // Verify all registered
    assert.Equal(t, len(plugins), len(manager.GetPlugins()))
    
    // Test concurrent access
    done := make(chan bool)
    go func() {
        for i := 0; i < 100; i++ {
            manager.GetPlugin(fmt.Sprintf("plugin-%d", i%10))
            manager.GetPluginsByType("BuildHook")
            time.Sleep(time.Microsecond)
        }
        done <- true
    }()
    
    go func() {
        for i := 0; i < 100; i++ {
            manager.GetPlugins()
            time.Sleep(time.Microsecond)
        }
        done <- true
    }()
    
    // Wait for completion
    <-done
    <-done
}

func TestManager_PanicRecovery(t *testing.T) {
    logger := testutil.NewTestLogger()
    config := &ManagerConfig{}
    manager := NewManager(config, logger)
    
    // Create plugin that panics
    panicPlugin := &PanicPlugin{
        MockPlugin: *testutil.NewMockPlugin("panic-plugin", "1.0.0"),
    }
    
    require.NoError(t, manager.registerPlugin(panicPlugin))
    
    ctx := context.Background()
    wailsConfig := &options.App{Name: "test-app"}
    require.NoError(t, manager.Initialize(ctx, wailsConfig))
    
    // Execute hook that causes panic
    buildCtx := &BuildContext{Context: ctx}
    
    // Should not panic
    err := manager.ExecuteHook("PreBuild", buildCtx)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "panic")
}

// PanicPlugin for testing panic recovery
type PanicPlugin struct {
    testutil.MockPlugin
}

func (p *PanicPlugin) PreBuild(ctx *BuildContext) error {
    panic("test panic")
}

func BenchmarkManager_GetPlugin(b *testing.B) {
    logger := testutil.NewTestLogger()
    config := &ManagerConfig{}
    manager := NewManager(config, logger)
    
    // Register 100 plugins
    for i := 0; i < 100; i++ {
        p := testutil.NewMockPlugin(fmt.Sprintf("plugin-%d", i), "1.0.0")
        manager.registerPlugin(p)
    }
    
    b.ResetTimer()
    
    b.Run("GetPlugin", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            manager.GetPlugin(fmt.Sprintf("plugin-%d", i%100))
        }
    })
    
    b.Run("GetPluginsByType", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            manager.GetPluginsByType("BuildHook")
        }
    })
}
```

### 3. Integration Tests (`integration_test.go`)
```go
package plugins_test

import (
    "context"
    "os"
    "path/filepath"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/wailsapp/wails/v2/pkg/plugins"
    "github.com/wailsapp/wails/v2/pkg/plugins/builtin/react"
    "github.com/wailsapp/wails/v2/pkg/plugins/wrf"
)

func TestPluginSystem_Integration(t *testing.T) {
    // Create test project
    tmpDir := t.TempDir()
    
    // Create project structure
    createTestProject(t, tmpDir)
    
    // Create configuration
    config := &plugins.ManagerConfig{
        Plugins: map[string]*plugins.PluginEntry{
            "basic-react": {
                Enabled: true,
                Config: map[string]interface{}{
                    "typesOutput": "frontend/src/types/generated.ts",
                    "hooksOutput": "frontend/src/hooks/generated",
                },
            },
        },
        LoadTimeout: 10 * time.Second,
    }
    
    logger := plugins.NewDefaultLogger()
    manager := plugins.NewManager(config, logger)
    
    // Register built-in plugins
    require.NoError(t, manager.RegisterPlugin(react.NewBasicReactPlugin()))
    
    // Initialize
    ctx := context.Background()
    wailsConfig := &options.App{
        Name: "test-app",
    }
    
    err := manager.Initialize(ctx, wailsConfig)
    require.NoError(t, err)
    
    // Run code generation
    genCtx := &plugins.GenerationContext{
        Context:     ctx,
        ProjectRoot: tmpDir,
        OutputDir:   tmpDir,
        Logger:      logger,
    }
    
    err = manager.ExecuteHook("GenerateCode", genCtx)
    require.NoError(t, err)
    
    // Verify generated files
    assert.FileExists(t, filepath.Join(tmpDir, "frontend/src/types/generated.ts"))
    assert.DirExists(t, filepath.Join(tmpDir, "frontend/src/hooks/generated"))
    
    // Shutdown
    err = manager.Shutdown(ctx)
    assert.NoError(t, err)
}

func createTestProject(t *testing.T, root string) {
    // Create directory structure
    dirs := []string{
        "app",
        "frontend/src/types",
        "frontend/src/hooks",
    }
    
    for _, dir := range dirs {
        err := os.MkdirAll(filepath.Join(root, dir), 0755)
        require.NoError(t, err)
    }
    
    // Create test Go service
    serviceCode := `
package app

import "context"

type TestService struct {
    ctx context.Context
}

func (s *TestService) GetUser(id string) (*User, error) {
    return &User{
        ID:   id,
        Name: "Test User",
    }, nil
}

type User struct {
    ID   string ` + "`json:\"id\"`" + `
    Name string ` + "`json:\"name\"`" + `
}
`
    
    err := os.WriteFile(filepath.Join(root, "app/service.go"), []byte(serviceCode), 0644)
    require.NoError(t, err)
}
```

### 4. React Plugin Tests (`builtin/react/react_test.go`)
```go
package react

import (
    "context"
    "strings"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/wailsapp/wails/v2/pkg/plugins"
)

func TestBasicReactPlugin_GenerateCode(t *testing.T) {
    plugin := NewBasicReactPlugin()
    
    // Initialize
    ctx := context.Background()
    config := &plugins.PluginConfig{
        ProjectRoot: t.TempDir(),
        Config: map[string]interface{}{
            "typesOutput": "types.ts",
            "hooksOutput": "hooks",
        },
        Logger: plugins.NewDefaultLogger(),
    }
    
    err := plugin.Initialize(ctx, config)
    require.NoError(t, err)
    
    // Create test service
    createTestService(t, config.ProjectRoot)
    
    // Generate code
    genCtx := &plugins.GenerationContext{
        Context:     ctx,
        ProjectRoot: config.ProjectRoot,
        Logger:      config.Logger,
    }
    
    files, err := plugin.GenerateCode(genCtx)
    require.NoError(t, err)
    assert.NotEmpty(t, files)
    
    // Check generated types
    var typesFile *plugins.GeneratedFile
    for _, f := range files {
        if strings.HasSuffix(f.Path, "types.ts") {
            typesFile = f
            break
        }
    }
    
    require.NotNil(t, typesFile)
    content := string(typesFile.Content)
    
    // Verify content
    assert.Contains(t, content, "export interface TestServiceService")
    assert.Contains(t, content, "getUser(id: string): Promise<User>")
    assert.Contains(t, content, "export interface User")
}

func TestBasicReactPlugin_TypeMapping(t *testing.T) {
    tests := []struct {
        goType string
        tsType string
    }{
        {"string", "string"},
        {"int", "number"},
        {"int64", "number"},
        {"bool", "boolean"},
        {"[]byte", "Uint8Array"},
        {"time.Time", "Date"},
        {"map[string]interface{}", "Record<string, any>"},
        {"[]string", "string[]"},
        {"*User", "User | null"},
    }
    
    plugin := &BasicReactPlugin{}
    
    for _, tt := range tests {
        t.Run(tt.goType, func(t *testing.T) {
            result := plugin.mapGoTypeToTS(tt.goType)
            assert.Equal(t, tt.tsType, result)
        })
    }
}
```

### 5. WRF Plugin Tests (`wrf/wrf_test.go`)
```go
package wrf

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestParseAnnotations(t *testing.T) {
    tests := []struct {
        name     string
        comments []string
        want     []Annotation
    }{
        {
            name: "query annotation",
            comments: []string{
                "// @wrf:query(cache: \"5m\", events: [\"user:updated\"])",
            },
            want: []Annotation{
                {
                    Type: "query",
                    Params: map[string]string{
                        "cache":  "5m",
                        "events": "[\"user:updated\"]",
                    },
                },
            },
        },
        {
            name: "mutation annotation",
            comments: []string{
                "// @wrf:mutation(optimistic: true, invalidate: [[\"users\"]])",
            },
            want: []Annotation{
                {
                    Type: "mutation",
                    Params: map[string]string{
                        "optimistic":  "true",
                        "invalidate": "[[\"users\"]]",
                    },
                },
            },
        },
        {
            name: "multiple annotations",
            comments: []string{
                "// GetUser retrieves a user",
                "// @wrf:query(cache: \"5m\")",
                "// @wrf:event(\"user:get\")",
            },
            want: []Annotation{
                {
                    Type:   "query",
                    Params: map[string]string{"cache": "5m"},
                },
                {
                    Type:   "event",
                    Params: map[string]string{"0": "\"user:get\""},
                },
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := ParseAnnotations(tt.comments)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### 6. Test Fixtures
Create test fixtures in `/v2/test/fixtures/plugins/`:
- Sample Go services with various patterns
- Expected generated TypeScript files
- Configuration examples
- Error cases

## Testing Strategy
- Unit tests for each component
- Integration tests for full workflows
- Performance benchmarks for critical paths
- Race condition tests for concurrent operations
- Mock implementations for isolation
- Table-driven tests for comprehensive coverage

## CI Integration
Update CI configuration to:
- Run plugin tests separately
- Generate coverage reports
- Run benchmarks on performance-critical code
- Check for race conditions

## Next Steps
With testing infrastructure complete, create documentation and examples.