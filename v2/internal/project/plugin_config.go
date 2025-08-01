package project

// PluginSystemConfig configures the plugin system in wails.json
type PluginSystemConfig struct {
	// Enable/disable entire plugin system
	Enabled bool `json:"enabled"`
	
	// Plugin configurations by name
	Plugins map[string]map[string]interface{} `json:"plugins,omitempty"`
	
	// Required - fail build if plugins fail to initialize
	Required bool `json:"required,omitempty"`
}