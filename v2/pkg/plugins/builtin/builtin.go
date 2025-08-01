package builtin

import (
	"github.com/wailsapp/wails/v2/pkg/plugins"
	"github.com/wailsapp/wails/v2/pkg/plugins/wrf"
)

// RegisterBuiltinPlugins registers all built-in plugins with the manager
func RegisterBuiltinPlugins(manager *plugins.Manager) {
	// Register advanced WRF plugin
	manager.RegisterPlugin(wrf.NewWRFPlugin())
}

// GetBuiltinPlugins returns a list of all built-in plugins
func GetBuiltinPlugins() []plugins.Plugin {
	return []plugins.Plugin{
		wrf.NewWRFPlugin(),
	}
}