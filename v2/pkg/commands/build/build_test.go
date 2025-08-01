package build

import (
	"testing"
	"time"
	
	"github.com/stretchr/testify/assert"
	"github.com/wailsapp/wails/v2/internal/project"
	"github.com/wailsapp/wails/v2/pkg/clilogger"
	"github.com/wailsapp/wails/v2/pkg/plugins"
)

func TestBuildModeConversion(t *testing.T) {
	tests := []struct {
		mode     Mode
		expected string
	}{
		{Dev, "development"},
		{Production, "production"},
		{Debug, "debug"},
		{Mode(99), "unknown"},
	}

	for _, tt := range tests {
		opts := &Options{Mode: tt.mode}
		result := getBuildMode(opts)
		assert.Equal(t, tt.expected, result)
	}
}

func TestExtractProjectConfig(t *testing.T) {
	proj := &project.Project{
		Name:           "test-app",
		OutputFilename: "test",
		FrontendDir:    "frontend",
		WailsJSDir:     "frontend/src/wailsjs",
		BuildTags:      "desktop production",
	}

	config := extractProjectConfig(proj)

	assert.Equal(t, "test-app", config["name"])
	assert.Equal(t, "test", config["outputfilename"])
	assert.Equal(t, "frontend", config["frontend:dir"])
	assert.Equal(t, "frontend/src/wailsjs", config["wailsjsdir"])
}

func TestWriteGeneratedFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		file      *plugins.GeneratedFile
		expectErr bool
	}{
		{
			name: "successful write",
			file: &plugins.GeneratedFile{
				Path:      "test.ts",
				Content:   []byte("export const test = true;"),
				Mode:      0644,
				Overwrite: true,
			},
			expectErr: false,
		},
		{
			name: "skip existing without overwrite",
			file: &plugins.GeneratedFile{
				Path:      "test.ts",
				Content:   []byte("new content"),
				Overwrite: false,
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := writeGeneratedFile(tmpDir, tt.file)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// MockPlugin for testing
type MockPlugin struct {
	name     string
	execPre  func(*plugins.BuildContext) error
	execPost func(*plugins.BuildContext) error
	genCode  func(*plugins.GenerationContext) ([]*plugins.GeneratedFile, error)
}

func (m *MockPlugin) Name() string    { return m.name }
func (m *MockPlugin) Version() string { return "1.0.0" }
func (m *MockPlugin) PreBuild(ctx *plugins.BuildContext) error {
	if m.execPre != nil {
		return m.execPre(ctx)
	}
	return nil
}
func (m *MockPlugin) PostBuild(ctx *plugins.BuildContext) error {
	if m.execPost != nil {
		return m.execPost(ctx)
	}
	return nil
}
func (m *MockPlugin) GenerateCode(ctx *plugins.GenerationContext) ([]*plugins.GeneratedFile, error) {
	if m.genCode != nil {
		return m.genCode(ctx)
	}
	return nil, nil
}

func TestPluginHooksExecution(t *testing.T) {
	logger := clilogger.New(nil)
	logger.Mute(true)

	buildCtx := &plugins.BuildContext{
		ProjectRoot: "/test",
		BuildMode:   "production",
		Platform:    "linux",
		Arch:        "amd64",
		Logger:      logger,
		StartTime:   time.Now(),
	}

	t.Run("pre-build hooks", func(t *testing.T) {
		called := false
		plugin := &MockPlugin{
			name: "test",
			execPre: func(ctx *plugins.BuildContext) error {
				called = true
				assert.Equal(t, "/test", ctx.ProjectRoot)
				return nil
			},
		}

		manager := plugins.NewManager()
		manager.RegisterPlugin(plugin)

		err := executePreBuildHooks(manager, buildCtx)
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("post-build hooks with errors", func(t *testing.T) {
		plugin := &MockPlugin{
			name: "test",
			execPost: func(ctx *plugins.BuildContext) error {
				return assert.AnError
			},
		}

		manager := plugins.NewManager()
		manager.RegisterPlugin(plugin)

		err := executePostBuildHooks(manager, buildCtx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "post-build errors")
	})
}