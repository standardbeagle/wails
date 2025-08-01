package wrf

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wailsapp/wails/v2/pkg/clilogger"
	"github.com/wailsapp/wails/v2/pkg/plugins"
)

func TestWRFPlugin_Initialize(t *testing.T) {
	plugin := NewWRFPlugin().(*WRFPlugin)
	logger := clilogger.New(nil)
	logger.Mute(true)

	config := &plugins.PluginConfig{
		ProjectRoot: "/test/project",
		Logger:      logger,
		Config: map[string]interface{}{
			"go": map[string]interface{}{
				"sourcePaths": []interface{}{"./app/**/*.go"},
				"watchMode":   true,
			},
			"generation": map[string]interface{}{
				"hooks": map[string]interface{}{
					"output":    "src/hooks",
					"framework": "react-query",
				},
				"types": map[string]interface{}{
					"output":     "src/types.ts",
					"validation": "zod",
				},
			},
		},
	}

	err := plugin.Initialize(context.Background(), config)
	require.NoError(t, err)

	assert.Equal(t, "wrf", plugin.Name())
	assert.Equal(t, "2.0.0", plugin.Version())
	assert.Equal(t, "/test/project", plugin.projectRoot)
	assert.Equal(t, "react-query", plugin.config.Generation.Hooks.Framework)
	assert.Equal(t, "zod", plugin.config.Generation.Types.Validation)
}

func TestWRFPlugin_Health(t *testing.T) {
	plugin := &WRFPlugin{}

	// Should fail without initialization
	err := plugin.Health()
	assert.Error(t, err)

	// Should pass after initialization
	plugin.config = DefaultConfig()
	err = plugin.Health()
	assert.NoError(t, err)
}

func TestAnnotationParsing(t *testing.T) {
	comments := []string{
		"GetUser retrieves a user by ID",
		"@wrf:query(cache: \"5m\", events: \"user:updated,user:deleted\")",
		"Returns user data or error",
	}

	annotations := ParseAnnotations(comments)
	require.Len(t, annotations, 1)

	ann := annotations[0]
	assert.Equal(t, "query", ann.Type)
	assert.Equal(t, "5m", ann.Params["cache"])
	assert.Equal(t, "user:updated,user:deleted", ann.Params["events"])
}

func TestQueryAnnotationExtraction(t *testing.T) {
	annotations := []Annotation{
		{
			Type: "query",
			Params: map[string]string{
				"cache":     "5m",
				"staleTime": "1m",
				"events":    "user:updated,user:deleted",
				"enabled":   "true",
			},
		},
	}

	queryAnn := ExtractQueryAnnotation(annotations)
	require.NotNil(t, queryAnn)

	assert.Equal(t, 5*time.Minute, queryAnn.Cache)
	assert.Equal(t, 1*time.Minute, queryAnn.StaleTime)
	assert.Equal(t, []string{"user:updated", "user:deleted"}, queryAnn.Events)
	assert.True(t, queryAnn.Enabled)
}

func TestMutationAnnotationExtraction(t *testing.T) {
	annotations := []Annotation{
		{
			Type: "mutation",
			Params: map[string]string{
				"optimistic":  "true",
				"invalidate":  "UserService:GetUser,UserService:ListUsers",
				"events":      "user:updated",
				"retryCount":  "5",
			},
		},
	}

	mutationAnn := ExtractMutationAnnotation(annotations)
	require.NotNil(t, mutationAnn)

	assert.True(t, mutationAnn.Optimistic)
	assert.Equal(t, []string{"UserService:GetUser", "UserService:ListUsers"}, mutationAnn.Invalidate)
	assert.Equal(t, []string{"user:updated"}, mutationAnn.Events)
	assert.Equal(t, 5, mutationAnn.RetryCount)
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		expectErr bool
	}{
		{
			name:      "valid config",
			config:    DefaultConfig(),
			expectErr: false,
		},
		{
			name: "empty source paths",
			config: &Config{
				Go: GoConfig{
					SourcePaths: []string{},
				},
				Generation: GenerationConfig{
					Hooks: HooksConfig{Output: "hooks"},
					Types: TypesConfig{Output: "types.ts"},
				},
			},
			expectErr: true,
		},
		{
			name: "invalid framework",
			config: &Config{
				Go: GoConfig{
					SourcePaths: []string{"./app/*.go"},
				},
				Generation: GenerationConfig{
					Hooks: HooksConfig{
						Output:    "hooks",
						Framework: "invalid-framework",
					},
					Types: TypesConfig{Output: "types.ts"},
				},
			},
			expectErr: true,
		},
		{
			name: "invalid validation",
			config: &Config{
				Go: GoConfig{
					SourcePaths: []string{"./app/*.go"},
				},
				Generation: GenerationConfig{
					Hooks: HooksConfig{
						Output:    "hooks",
						Framework: "react-query",
					},
					Types: TypesConfig{
						Output:     "types.ts",
						Validation: "invalid-validation",
					},
				},
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

func TestGoTypeToTS(t *testing.T) {
	generator := &Generator{config: DefaultConfig()}

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
		{"CustomType", "CustomType"},
	}

	for _, tt := range tests {
		result := generator.goTypeToTS(tt.goType)
		assert.Equal(t, tt.expected, result, "Converting %s", tt.goType)
	}
}

func TestGoTypeToZod(t *testing.T) {
	generator := &Generator{config: DefaultConfig()}

	tests := []struct {
		goType   string
		expected string
	}{
		{"string", "z.string()"},
		{"int", "z.number().int()"},
		{"float64", "z.number()"},
		{"bool", "z.boolean()"},
		{"[]byte", "z.instanceof(Uint8Array)"},
		{"interface{}", "z.any()"},
		{"[]string", "z.array(z.string())"},
		{"*string", "z.string().nullable()"},
		{"map[string]interface{}", "z.record(z.any())"},
		{"time.Time", "z.date()"},
	}

	for _, tt := range tests {
		result := generator.goTypeToZod(tt.goType)
		assert.Equal(t, tt.expected, result, "Converting %s", tt.goType)
	}
}

func TestFieldOptionalDetection(t *testing.T) {
	generator := &Generator{config: DefaultConfig()}

	tests := []struct {
		name     string
		field    Field
		expected bool
	}{
		{
			name: "required field",
			field: Field{
				Name: "Name",
				Type: "string",
				Tag:  &StructTag{Raw: `json:"name"`},
			},
			expected: false,
		},
		{
			name: "omitempty field",
			field: Field{
				Name: "Email",
				Type: "string",
				Tag:  &StructTag{Raw: `json:"email,omitempty"`},
			},
			expected: true,
		},
		{
			name: "pointer field",
			field: Field{
				Name: "Age",
				Type: "*int",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.isFieldOptional(tt.field)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidationExtraction(t *testing.T) {
	generator := &Generator{config: DefaultConfig()}

	tests := []struct {
		name     string
		field    Field
		expected string
	}{
		{
			name: "no validation",
			field: Field{
				Tag: &StructTag{Raw: `json:"name"`},
			},
			expected: "",
		},
		{
			name: "min validation",
			field: Field{
				Tag: &StructTag{Raw: `wrf:"min:5" json:"name"`},
			},
			expected: ".min(5)",
		},
		{
			name: "email validation",
			field: Field{
				Tag: &StructTag{Raw: `wrf:"validation:email" json:"email"`},
			},
			expected: ".email()",
		},
		{
			name: "multiple validations",
			field: Field{
				Tag: &StructTag{Raw: `wrf:"min:1,max:100,validation:url" json:"url"`},
			},
			expected: ".min(1).max(100).url()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.extractValidation(tt.field)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCodeGenerationIntegration(t *testing.T) {
	// Create a temporary project structure
	tmpDir := t.TempDir()
	appDir := filepath.Join(tmpDir, "app")
	require.NoError(t, os.MkdirAll(appDir, 0755))

	// Create a sample Go file with WRF annotations
	goFile := filepath.Join(appDir, "user.go")
	goContent := `package app

import "context"

// User represents a user in the system
type User struct {
	ID    int    ` + "`json:\"id\"`" + `
	Name  string ` + "`json:\"name\"`" + `
	Email string ` + "`json:\"email,omitempty\" wrf:\"validation:email\"`" + `
}

// UserService manages users
type UserService struct {}

// GetUser retrieves a user by ID
// @wrf:query(cache: "5m", events: "user:updated")
func (u *UserService) GetUser(ctx context.Context, id int) (*User, error) {
	return nil, nil
}

// UpdateUser updates a user
// @wrf:mutation(optimistic: true, invalidate: "UserService:GetUser", events: "user:updated")
func (u *UserService) UpdateUser(ctx context.Context, user *User) (*User, error) {
	return nil, nil
}
`
	require.NoError(t, os.WriteFile(goFile, []byte(goContent), 0644))

	// Initialize plugin
	plugin := NewWRFPlugin().(*WRFPlugin)
	logger := clilogger.New(nil)
	logger.Mute(true)

	config := &plugins.PluginConfig{
		ProjectRoot: tmpDir,
		Logger:      logger,
		Config: map[string]interface{}{
			"go": map[string]interface{}{
				"sourcePaths": []interface{}{filepath.Join("app", "*.go")},
			},
			"generation": map[string]interface{}{
				"hooks": map[string]interface{}{
					"output":    "hooks",
					"framework": "react-query",
				},
				"types": map[string]interface{}{
					"output":     "types.ts",
					"validation": "zod",
				},
			},
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

	// Should generate multiple files
	assert.GreaterOrEqual(t, len(files), 3)

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
	assert.Contains(t, typesContent, "export interface User")
	assert.Contains(t, typesContent, "id: number")
	assert.Contains(t, typesContent, "email?: string")

	// Check hooks file
	var hooksFile *plugins.GeneratedFile
	for _, file := range files {
		if strings.Contains(file.Path, "UserServiceHooks.ts") {
			hooksFile = file
			break
		}
	}
	require.NotNil(t, hooksFile)

	hooksContent := string(hooksFile.Content)
	assert.Contains(t, hooksContent, "useGetUser")
	assert.Contains(t, hooksContent, "useUpdateUser")
	assert.Contains(t, hooksContent, "useQuery")
	assert.Contains(t, hooksContent, "useMutation")

	// Check schema file (if Zod enabled)
	var schemaFile *plugins.GeneratedFile
	for _, file := range files {
		if strings.Contains(file.Path, "schemas.ts") {
			schemaFile = file
			break
		}
	}
	if schemaFile != nil {
		schemaContent := string(schemaFile.Content)
		assert.Contains(t, schemaContent, "UserSchema")
		assert.Contains(t, schemaContent, "z.object")
		assert.Contains(t, schemaContent, ".email()")
	}
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
		{"XMLHttpRequest", "xMLHttpRequest"},
	}

	for _, tt := range tests {
		result := lowerFirst(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		hasError bool
	}{
		{"5", 5 * time.Second, false},
		{"300", 300 * time.Second, false},
		{"5m", 5 * time.Minute, false},
		{"1h", 1 * time.Hour, false},
		{"30s", 30 * time.Second, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		result, err := parseDuration(tt.input)
		if tt.hasError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		}
	}
}

func TestParseStringArray(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"", nil},
		{"single", []string{"single"}},
		{"one,two,three", []string{"one", "two", "three"}},
		{"one, two, three ", []string{"one", "two", "three"}},
		{" , , ", nil},
	}

	for _, tt := range tests {
		result := parseStringArray(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}