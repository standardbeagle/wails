package plugins

import (
	"context"
	"testing"
)

// MockPlugin implements the Plugin interface for testing
type MockPlugin struct {
	name        string
	version     string
	description string
	author      string
	initialized bool
	shutdown    bool
	healthError error
}

func NewMockPlugin(name string) *MockPlugin {
	return &MockPlugin{
		name:        name,
		version:     "1.0.0",
		description: "Mock plugin for testing",
		author:      "Test Author",
	}
}

func (m *MockPlugin) Name() string        { return m.name }
func (m *MockPlugin) Version() string     { return m.version }
func (m *MockPlugin) Description() string { return m.description }
func (m *MockPlugin) Author() string      { return m.author }

func (m *MockPlugin) Initialize(ctx context.Context, config *PluginConfig) error {
	m.initialized = true
	return nil
}

func (m *MockPlugin) Shutdown(ctx context.Context) error {
	m.shutdown = true
	return nil
}

func (m *MockPlugin) Health() error {
	return m.healthError
}

// MockBuildHookPlugin implements both Plugin and BuildHook interfaces
type MockBuildHookPlugin struct {
	*MockPlugin
	preBuildCalled  bool
	postBuildCalled bool
}

func NewMockBuildHookPlugin(name string) *MockBuildHookPlugin {
	return &MockBuildHookPlugin{
		MockPlugin: NewMockPlugin(name),
	}
}

func (m *MockBuildHookPlugin) PreBuild(ctx *BuildContext) error {
	m.preBuildCalled = true
	return nil
}

func (m *MockBuildHookPlugin) PostBuild(ctx *BuildContext) error {
	m.postBuildCalled = true
	return nil
}

// MockCodeGeneratorPlugin implements both Plugin and CodeGenerator interfaces
type MockCodeGeneratorPlugin struct {
	*MockPlugin
	languages []string
}

func NewMockCodeGeneratorPlugin(name string) *MockCodeGeneratorPlugin {
	return &MockCodeGeneratorPlugin{
		MockPlugin: NewMockPlugin(name),
		languages:  []string{"go", "typescript", "javascript"},
	}
}

func (m *MockCodeGeneratorPlugin) GenerateCode(ctx *GenerationContext) ([]*GeneratedFile, error) {
	return []*GeneratedFile{
		{
			Path:     "generated.go",
			Content:  []byte("package main\n\nfunc main() {}\n"),
			Language: "go",
		},
	}, nil
}

func (m *MockCodeGeneratorPlugin) SupportedLanguages() []string {
	return m.languages
}

func TestPluginInterface(t *testing.T) {
	plugin := NewMockPlugin("test-plugin")

	// Test metadata methods
	if plugin.Name() != "test-plugin" {
		t.Errorf("Expected name 'test-plugin', got '%s'", plugin.Name())
	}

	if plugin.Version() != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", plugin.Version())
	}

	if plugin.Description() != "Mock plugin for testing" {
		t.Errorf("Expected description 'Mock plugin for testing', got '%s'", plugin.Description())
	}

	if plugin.Author() != "Test Author" {
		t.Errorf("Expected author 'Test Author', got '%s'", plugin.Author())
	}

	// Test lifecycle methods
	ctx := context.Background()
	config := &PluginConfig{
		ProjectRoot: "/test/project",
		ProjectName: "test-project",
	}

	err := plugin.Initialize(ctx, config)
	if err != nil {
		t.Errorf("Initialize failed: %v", err)
	}

	if !plugin.initialized {
		t.Error("Plugin was not marked as initialized")
	}

	err = plugin.Health()
	if err != nil {
		t.Errorf("Health check failed: %v", err)
	}

	err = plugin.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	if !plugin.shutdown {
		t.Error("Plugin was not marked as shutdown")
	}
}

func TestBuildHookInterface(t *testing.T) {
	plugin := NewMockBuildHookPlugin("build-hook-plugin")

	// Verify it implements Plugin interface
	var p Plugin = plugin
	if p.Name() != "build-hook-plugin" {
		t.Errorf("Expected name 'build-hook-plugin', got '%s'", p.Name())
	}

	// Verify it implements BuildHook interface
	var buildHook BuildHook = plugin

	buildContext := &BuildContext{
		Context:     context.Background(),
		ProjectRoot: "/test/project",
		ProjectName: "test-project",
		BuildMode:   "development",
	}

	err := buildHook.PreBuild(buildContext)
	if err != nil {
		t.Errorf("PreBuild failed: %v", err)
	}

	if !plugin.preBuildCalled {
		t.Error("PreBuild was not called")
	}

	err = buildHook.PostBuild(buildContext)
	if err != nil {
		t.Errorf("PostBuild failed: %v", err)
	}

	if !plugin.postBuildCalled {
		t.Error("PostBuild was not called")
	}
}

func TestCodeGeneratorInterface(t *testing.T) {
	plugin := NewMockCodeGeneratorPlugin("code-gen-plugin")

	// Verify it implements Plugin interface
	var p Plugin = plugin
	if p.Name() != "code-gen-plugin" {
		t.Errorf("Expected name 'code-gen-plugin', got '%s'", p.Name())
	}

	// Verify it implements CodeGenerator interface
	var codeGen CodeGenerator = plugin

	languages := codeGen.SupportedLanguages()
	expectedLanguages := []string{"go", "typescript", "javascript"}
	
	if len(languages) != len(expectedLanguages) {
		t.Errorf("Expected %d languages, got %d", len(expectedLanguages), len(languages))
	}

	for i, lang := range languages {
		if lang != expectedLanguages[i] {
			t.Errorf("Expected language '%s', got '%s'", expectedLanguages[i], lang)
		}
	}

	genContext := &GenerationContext{
		Context:     context.Background(),
		ProjectRoot: "/test/project",
		ProjectName: "test-project",
		Language:    "go",
	}

	files, err := codeGen.GenerateCode(genContext)
	if err != nil {
		t.Errorf("GenerateCode failed: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 generated file, got %d", len(files))
	}

	if files[0].Path != "generated.go" {
		t.Errorf("Expected file path 'generated.go', got '%s'", files[0].Path)
	}

	if files[0].Language != "go" {
		t.Errorf("Expected language 'go', got '%s'", files[0].Language)
	}
}

func TestInterfaceCompatibility(t *testing.T) {
	// Test that plugins can implement multiple interfaces
	buildHookPlugin := NewMockBuildHookPlugin("multi-interface-plugin")

	// Should implement Plugin
	var plugin Plugin = buildHookPlugin
	if plugin == nil {
		t.Error("Plugin interface not implemented")
	}

	// Should implement BuildHook
	var buildHook BuildHook = buildHookPlugin
	if buildHook == nil {
		t.Error("BuildHook interface not implemented")
	}

	// Test type assertions
	if _, ok := plugin.(BuildHook); !ok {
		t.Error("Plugin should be assertable to BuildHook interface")
	}

	// Test that basic Plugin functionality still works
	if plugin.Name() != "multi-interface-plugin" {
		t.Errorf("Expected name 'multi-interface-plugin', got '%s'", plugin.Name())
	}
}

func TestPluginState(t *testing.T) {
	tests := []struct {
		state    PluginState
		expected string
	}{
		{StateUninitialized, "uninitialized"},
		{StateInitializing, "initializing"},
		{StateReady, "ready"},
		{StateShuttingDown, "shutting_down"},
		{StateStopped, "stopped"},
		{StateError, "error"},
		{PluginState(999), "unknown"},
	}

	for _, test := range tests {
		if test.state.String() != test.expected {
			t.Errorf("Expected state string '%s', got '%s'", test.expected, test.state.String())
		}
	}
}

func TestValidStateTransitions(t *testing.T) {
	validTransitions := map[PluginState][]PluginState{
		StateUninitialized: {StateInitializing, StateError},
		StateInitializing:  {StateReady, StateError},
		StateReady:         {StateShuttingDown, StateError},
		StateShuttingDown:  {StateStopped, StateError},
		StateStopped:       {StateInitializing, StateError},
		StateError:         {StateInitializing, StateStopped},
	}

	for fromState, validToStates := range validTransitions {
		for _, toState := range validToStates {
			if !fromState.IsValidTransition(toState) {
				t.Errorf("Transition from %s to %s should be valid", fromState, toState)
			}
		}

		// Test some invalid transitions
		allStates := []PluginState{
			StateUninitialized, StateInitializing, StateReady,
			StateShuttingDown, StateStopped, StateError,
		}

		for _, toState := range allStates {
			isValid := false
			for _, validState := range validToStates {
				if toState == validState {
					isValid = true
					break
				}
			}

			if !isValid && fromState.IsValidTransition(toState) {
				t.Errorf("Transition from %s to %s should be invalid", fromState, toState)
			}
		}
	}
}