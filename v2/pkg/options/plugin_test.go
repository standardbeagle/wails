package options

import (
	"encoding/json"
	"testing"
	"time"
)

func TestPluginSystemConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *PluginSystemConfig
		wantErr bool
	}{
		{
			name: "Valid config with plugins",
			config: &PluginSystemConfig{
				Enabled:     true,
				LoadTimeout: 10 * time.Second,
				Plugins: map[string]*PluginConfig{
					"test-plugin": {
						Enabled: true,
						Version: "1.0.0",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid empty config",
			config: &PluginSystemConfig{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "Negative load timeout",
			config: &PluginSystemConfig{
				Enabled:     true,
				LoadTimeout: -5 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "Empty plugin name",
			config: &PluginSystemConfig{
				Enabled: true,
				Plugins: map[string]*PluginConfig{
					"": {
						Enabled: true,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Nil plugin config",
			config: &PluginSystemConfig{
				Enabled: true,
				Plugins: map[string]*PluginConfig{
					"test-plugin": nil,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("PluginSystemConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPluginSystemConfig_SetDefaults(t *testing.T) {
	config := &PluginSystemConfig{}
	config.SetDefaults()

	if config.LoadTimeout != 30*time.Second {
		t.Errorf("Expected LoadTimeout to be 30s, got %v", config.LoadTimeout)
	}

	expectedPaths := []string{"./.wails/plugins", "~/.wails/plugins"}
	if len(config.DiscoveryPaths) != len(expectedPaths) {
		t.Errorf("Expected %d discovery paths, got %d", len(expectedPaths), len(config.DiscoveryPaths))
	}

	for i, path := range expectedPaths {
		if config.DiscoveryPaths[i] != path {
			t.Errorf("Expected discovery path %s, got %s", path, config.DiscoveryPaths[i])
		}
	}
}

func TestPluginSystemConfig_JSONMarshaling(t *testing.T) {
	config := &PluginSystemConfig{
		Enabled:        true,
		DiscoveryPaths: []string{"./plugins", "~/.wails/plugins"},
		LoadTimeout:    30 * time.Second,
		Plugins: map[string]*PluginConfig{
			"wrf": {
				Enabled: true,
				Version: "^1.0.0",
				Config: map[string]interface{}{
					"go": map[string]interface{}{
						"sourcePaths": []interface{}{"./app/**/*.go"},
					},
					"generation": map[string]interface{}{
						"hooks": map[string]interface{}{
							"output": "frontend/src/hooks/generated",
						},
					},
				},
			},
			"basic-vue": {
				Enabled: false,
			},
		},
	}

	// Test marshaling
	data, err := json.Marshal(config)
	if err != nil {
		t.Errorf("Failed to marshal config: %v", err)
	}

	// Test unmarshaling
	var unmarshaled PluginSystemConfig
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal config: %v", err)
	}

	// Test specific values
	if unmarshaled.Enabled != config.Enabled {
		t.Errorf("Expected Enabled to be %v, got %v", config.Enabled, unmarshaled.Enabled)
	}

	if len(unmarshaled.Plugins) != len(config.Plugins) {
		t.Errorf("Expected %d plugins, got %d", len(config.Plugins), len(unmarshaled.Plugins))
	}

	wrfPlugin := unmarshaled.Plugins["wrf"]
	if wrfPlugin == nil {
		t.Error("Expected wrf plugin to be present")
	} else {
		if wrfPlugin.Enabled != true {
			t.Error("Expected wrf plugin to be enabled")
		}
		if wrfPlugin.Version != "^1.0.0" {
			t.Errorf("Expected wrf plugin version to be '^1.0.0', got %s", wrfPlugin.Version)
		}
	}
}

func TestApp_Validate_WithPlugins(t *testing.T) {
	tests := []struct {
		name    string
		app     *App
		wantErr bool
	}{
		{
			name: "Valid app with no plugins",
			app:  &App{},
			wantErr: false,
		},
		{
			name: "Valid app with valid plugins",
			app: &App{
				Plugins: &PluginSystemConfig{
					Enabled: true,
					Plugins: map[string]*PluginConfig{
						"test": {
							Enabled: true,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid app with invalid plugins",
			app: &App{
				Plugins: &PluginSystemConfig{
					Enabled:     true,
					LoadTimeout: -5 * time.Second,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.app.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("App.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestApp_JSONMarshaling_BackwardCompatibility(t *testing.T) {
	// Test that apps without plugins field still work
	appJSON := `{"title": "Test App", "width": 800, "height": 600}`
	
	var app App
	err := json.Unmarshal([]byte(appJSON), &app)
	if err != nil {
		t.Errorf("Failed to unmarshal app without plugins: %v", err)
	}

	if app.Title != "Test App" {
		t.Errorf("Expected title 'Test App', got %s", app.Title)
	}

	if app.Plugins != nil {
		t.Error("Expected Plugins to be nil for backward compatibility")
	}

	// Test validation works
	err = app.Validate()
	if err != nil {
		t.Errorf("Validation should pass for app without plugins: %v", err)
	}
}

func TestApp_JSONMarshaling_WithPlugins(t *testing.T) {
	// Test full app configuration with plugins
	appJSON := `{
		"title": "Test App",
		"width": 800,
		"height": 600,
		"plugins": {
			"enabled": true,
			"discoveryPaths": ["./plugins"],
			"loadTimeout": "30s",
			"plugins": {
				"wrf": {
					"enabled": true,
					"version": "^1.0.0",
					"config": {
						"go": {
							"sourcePaths": ["./app/**/*.go"]
						}
					}
				}
			}
		}
	}`

	var app App
	err := json.Unmarshal([]byte(appJSON), &app)
	if err != nil {
		t.Errorf("Failed to unmarshal app with plugins: %v", err)
	}

	if app.Title != "Test App" {
		t.Errorf("Expected title 'Test App', got %s", app.Title)
	}

	if app.Plugins == nil {
		t.Error("Expected plugins configuration to be present")
		return
	}

	if !app.Plugins.Enabled {
		t.Error("Expected plugins to be enabled")
	}

	if len(app.Plugins.DiscoveryPaths) != 1 || app.Plugins.DiscoveryPaths[0] != "./plugins" {
		t.Errorf("Expected discovery paths to be ['./plugins'], got %v", app.Plugins.DiscoveryPaths)
	}

	wrfPlugin := app.Plugins.Plugins["wrf"]
	if wrfPlugin == nil {
		t.Error("Expected wrf plugin to be present")
	} else {
		if !wrfPlugin.Enabled {
			t.Error("Expected wrf plugin to be enabled")
		}
		if wrfPlugin.Version != "^1.0.0" {
			t.Errorf("Expected wrf plugin version to be '^1.0.0', got %s", wrfPlugin.Version)
		}
	}

	// Test validation
	err = app.Validate()
	if err != nil {
		t.Errorf("Validation should pass for valid app with plugins: %v", err)
	}
}

func TestMergeDefaults_WithPlugins(t *testing.T) {
	app := &App{
		Plugins: &PluginSystemConfig{
			Enabled: true,
		},
	}

	MergeDefaults(app)

	// Check that plugin defaults were applied
	if app.Plugins.LoadTimeout != 30*time.Second {
		t.Errorf("Expected plugin LoadTimeout to be 30s, got %v", app.Plugins.LoadTimeout)
	}

	expectedPaths := []string{"./.wails/plugins", "~/.wails/plugins"}
	if len(app.Plugins.DiscoveryPaths) != len(expectedPaths) {
		t.Errorf("Expected %d discovery paths, got %d", len(expectedPaths), len(app.Plugins.DiscoveryPaths))
	}
}

func TestMergeDefaults_WithoutPlugins(t *testing.T) {
	app := &App{}

	MergeDefaults(app)

	// Check that app defaults were applied
	if app.Width != 1024 {
		t.Errorf("Expected width to be 1024, got %d", app.Width)
	}

	if app.Height != 768 {
		t.Errorf("Expected height to be 768, got %d", app.Height)
	}

	// Plugins should remain nil
	if app.Plugins != nil {
		t.Error("Expected Plugins to remain nil when not set")
	}
}