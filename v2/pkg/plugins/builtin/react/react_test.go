package react

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wailsapp/wails/v2/pkg/clilogger"
	"github.com/wailsapp/wails/v2/pkg/plugins"
)

func TestBasicReactPlugin_Initialize(t *testing.T) {
	plugin := NewBasicReactPlugin().(*BasicReactPlugin)
	logger := clilogger.New(nil)
	logger.Mute(true)

	config := &plugins.PluginConfig{
		ProjectRoot: "/test/project",
		Logger:      logger,
		Config: map[string]interface{}{
			"typesOutput":     "src/types.ts",
			"hooksOutput":     "src/hooks",
			"includeComments": true,
		},
	}

	err := plugin.Initialize(context.Background(), config)
	require.NoError(t, err)

	assert.Equal(t, "basic-react", plugin.Name())
	assert.Equal(t, "1.0.0", plugin.Version())
	assert.Equal(t, "/test/project", plugin.projectRoot)
	assert.Equal(t, "src/types.ts", plugin.config.TypesOutput)
	assert.Equal(t, "src/hooks", plugin.config.HooksOutput)
	assert.True(t, plugin.config.IncludeComments)
}

func TestBasicReactPlugin_Health(t *testing.T) {
	plugin := &BasicReactPlugin{}

	// Should fail without initialization
	err := plugin.Health()
	assert.Error(t, err)

	// Should pass after initialization
	plugin.config = &Config{
		TypesOutput: "types.ts",
		HooksOutput: "hooks",
	}
	err = plugin.Health()
	assert.NoError(t, err)
}

func TestGoTypeToTS(t *testing.T) {
	plugin := &BasicReactPlugin{}

	tests := []struct {
		goType   string
		expected string
	}{
		{"string", "string"},
		{"int", "number"},
		{"int32", "number"},
		{"float64", "number"},
		{"bool", "boolean"},
		{"[]byte", "Uint8Array"},
		{"error", "Error"},
		{"interface{}", "any"},
		{"[]string", "string[]"},
		{"*string", "string | null"},
		{"map[string]interface{}", "Record<string, any>"},
		{"time.Time", "Date"},
		{"CustomType", "any"},
	}

	for _, tt := range tests {
		result := plugin.goTypeToTS(tt.goType)
		assert.Equal(t, tt.expected, result, "Converting %s", tt.goType)
	}
}

func TestFormatParams(t *testing.T) {
	plugin := &BasicReactPlugin{}

	tests := []struct {
		params   []Param
		expected string
	}{
		{
			params:   []Param{},
			expected: "",
		},
		{
			params:   []Param{{Name: "name", Type: "string"}},
			expected: "name: string",
		},
		{
			params: []Param{
				{Name: "name", Type: "string"},
				{Name: "age", Type: "int"},
			},
			expected: "name: string, age: number",
		},
	}

	for _, tt := range tests {
		result := plugin.formatParams(tt.params)
		assert.Equal(t, tt.expected, result)
	}
}

func TestFormatReturns(t *testing.T) {
	plugin := &BasicReactPlugin{}

	tests := []struct {
		results  []Result
		expected string
	}{
		{
			results:  []Result{},
			expected: "void",
		},
		{
			results:  []Result{{Type: "string"}},
			expected: "string",
		},
		{
			results: []Result{
				{Type: "string"},
				{Type: "error"},
			},
			expected: "[string, Error]",
		},
	}

	for _, tt := range tests {
		result := plugin.formatReturns(tt.results)
		assert.Equal(t, tt.expected, result)
	}
}

func TestGenerateTypes(t *testing.T) {
	plugin := &BasicReactPlugin{
		config: &Config{
			TypesOutput:     "types.ts",
			IncludeComments: true,
		},
	}

	services := []Service{
		{
			Name: "UserService",
			Doc:  "User management service",
			Methods: []Method{
				{
					Name: "GetUser",
					Doc:  "Get user by ID",
					Params: []Param{
						{Name: "id", Type: "int"},
					},
					Results: []Result{
						{Type: "User"},
						{Type: "error"},
					},
				},
			},
		},
	}

	file, err := plugin.generateTypes(services)
	require.NoError(t, err)

	content := string(file.Content)
	assert.Contains(t, content, "export interface UserServiceService")
	assert.Contains(t, content, "getUser(id: number): Promise<[any, Error]>")
	assert.Contains(t, content, "User management service")
	assert.Contains(t, content, "Get user by ID")
	assert.Contains(t, content, "window.go")
}

func TestGenerateHooks(t *testing.T) {
	plugin := &BasicReactPlugin{
		config: &Config{
			HooksOutput: "hooks",
		},
	}

	services := []Service{
		{
			Name: "UserService",
			Methods: []Method{
				{
					Name: "GetUser",
					Params: []Param{
						{Name: "id", Type: "int"},
					},
					Results: []Result{
						{Type: "User"},
					},
				},
			},
		},
	}

	files, err := plugin.generateHooks(services)
	require.NoError(t, err)
	require.Len(t, files, 1)

	content := string(files[0].Content)
	assert.Contains(t, content, "export function useGetUser()")
	assert.Contains(t, content, "useState")
	assert.Contains(t, content, "useCallback")
	assert.Contains(t, content, "window.go.UserService.GetUser")
	assert.Equal(t, "hooks/UserServiceHooks.ts", files[0].Path)
}

func TestGenerateCodeIntegration(t *testing.T) {
	// Create a temporary project structure
	tmpDir := t.TempDir()
	appDir := filepath.Join(tmpDir, "app")
	require.NoError(t, os.MkdirAll(appDir, 0755))

	// Create a sample Go file
	goFile := filepath.Join(appDir, "user.go")
	goContent := `package app

// UserService manages users
type UserService struct {}

// GetUser retrieves a user by ID
func (u *UserService) GetUser(id int) (string, error) {
	return "user", nil
}

// CreateUser creates a new user
func (u *UserService) CreateUser(name string, email string) (int, error) {
	return 1, nil
}
`
	require.NoError(t, os.WriteFile(goFile, []byte(goContent), 0644))

	// Initialize plugin
	plugin := NewBasicReactPlugin().(*BasicReactPlugin)
	logger := clilogger.New(nil)
	logger.Mute(true)

	config := &plugins.PluginConfig{
		ProjectRoot: tmpDir,
		Logger:      logger,
		Config: map[string]interface{}{
			"typesOutput": "types.ts",
			"hooksOutput": "hooks",
		},
	}

	require.NoError(t, plugin.Initialize(context.Background(), config))

	// Generate code
	ctx := &plugins.GenerationContext{
		ProjectRoot: tmpDir,
		Logger:      logger,
	}

	files, err := plugin.GenerateCode(ctx)
	require.NoError(t, err)

	// Should generate types, hooks, and index files
	assert.Len(t, files, 3)

	// Check types file
	var typesFile *plugins.GeneratedFile
	for _, file := range files {
		if strings.HasSuffix(file.Path, "types.ts") {
			typesFile = file
			break
		}
	}
	require.NotNil(t, typesFile)

	typesContent := string(typesFile.Content)
	assert.Contains(t, typesContent, "UserServiceService")
	assert.Contains(t, typesContent, "getUser")
	assert.Contains(t, typesContent, "createUser")
}

func TestLowerFirst(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"A", "a"},
		{"GetUser", "getUser"},
		{"alreadyLower", "alreadyLower"},
	}

	for _, tt := range tests {
		result := lowerFirst(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		expectErr bool
	}{
		{
			name: "valid config",
			config: Config{
				TypesOutput: "types.ts",
				HooksOutput: "hooks",
			},
			expectErr: false,
		},
		{
			name: "missing types output",
			config: Config{
				HooksOutput: "hooks",
			},
			expectErr: true,
		},
		{
			name: "missing hooks output",
			config: Config{
				TypesOutput: "types.ts",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}