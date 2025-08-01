package builtin

import (
	"github.com/wailsapp/wails/v2/pkg/plugins"
	"github.com/wailsapp/wails/v2/pkg/plugins/builtin/react"
)

// RegisterBuiltinPlugins registers all built-in plugins with the manager
func RegisterBuiltinPlugins(manager *plugins.Manager) {
	// Register basic React plugin
	manager.RegisterPlugin(react.NewBasicReactPlugin())
}

// GetBuiltinPlugins returns a list of all built-in plugins
func GetBuiltinPlugins() []plugins.Plugin {
	return []plugins.Plugin{
		react.NewBasicReactPlugin(),
	}
}