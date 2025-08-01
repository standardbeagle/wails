package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/leaanthony/clir"
	"github.com/pterm/pterm"
	"github.com/wailsapp/wails/v2/internal/project"
	"github.com/wailsapp/wails/v2/pkg/clilogger"
	"github.com/wailsapp/wails/v2/pkg/plugins"
)

// pluginList implements the 'wails plugin list' command
func pluginList(f *struct{}) error {
	// Create logger
	logger := clilogger.New(os.Stdout)

	// Print banner
	app.PrintBanner()

	pterm.DefaultSection.Println("Available Wails Plugins")

	// Try to load project configuration for context
	cwd, _ := os.Getwd()
	var projectConfig *project.Project

	// Check if we're in a Wails project
	if projectData, err := project.Load(cwd); err == nil {
		projectConfig = projectData
	}

	// Create discovery options
	discoveryOpts := plugins.DefaultDiscoveryOptions()

	// Discover available plugins without initializing
	result, err := plugins.DiscoverAvailablePlugins(discoveryOpts)
	if err != nil {
		return fmt.Errorf("plugin discovery failed: %w", err)
	}

	// Display built-in plugins
	if len(result.BuiltinPlugins) > 0 {
		fmt.Println()
		pterm.DefaultHeader.Println("Built-in Plugins:")
		
		registry := plugins.GetBuiltinRegistry()
		for _, name := range result.BuiltinPlugins {
			plugin, err := registry.CreateBuiltin(name)
			if err != nil {
				pterm.Warning.Printf("  - %s (failed to load: %v)\n", name, err)
				continue
			}

			status := "available"
			if projectConfig != nil {
				// TODO: Check if plugin is enabled in project config
				status = "available"
			}

			pterm.Printf("  - %s (v%s) - %s [%s]\n", 
				pterm.Green(plugin.Name()), 
				plugin.Version(), 
				plugin.Description(), 
				pterm.Blue(status))
		}
	}

	// Display external plugins
	if len(result.ExternalPlugins) > 0 {
		fmt.Println()
		pterm.DefaultHeader.Println("External Plugins:")
		
		for _, manifest := range result.ExternalPlugins {
			status := "available"
			if projectConfig != nil {
				// TODO: Check if plugin is enabled in project config
				status = "available"
			}

			pterm.Printf("  - %s (v%s) - %s [%s]\n", 
				pterm.Green(manifest.Name), 
				manifest.Version, 
				manifest.Description, 
				pterm.Blue(status))
		}
	}

	// Display discovery errors if any
	if len(result.Errors) > 0 {
		fmt.Println()
		pterm.Warning.Println("Discovery Issues:")
		for _, err := range result.Errors {
			pterm.Warning.Printf("  - %v\n", err)
		}
	}

	// Show total count
	total := len(result.BuiltinPlugins) + len(result.ExternalPlugins)
	if total == 0 {
		fmt.Println()
		pterm.Info.Println("No plugins found. Consider installing plugins or check your plugin directories.")
		
		// Show default plugin directories
		fmt.Println()
		pterm.Info.Println("Default plugin search directories:")
		for _, dir := range discoveryOpts.PluginDirs {
			pterm.Printf("  - %s\n", dir)
		}
	} else {
		fmt.Println()
		pterm.Success.Printf("Found %d plugin(s) total\n", total)
	}

	return nil
}

// pluginInfo implements the 'wails plugin info' command - shows usage for now
func pluginInfo(f *struct{}) error {
	// Print banner
	app.PrintBanner()

	pterm.Error.Println("Plugin name is required as an argument")
	pterm.Info.Println("Usage: wails plugin info <plugin-name>")
	pterm.Info.Println("Example: wails plugin info basic-react")
	
	// Show available plugins as help
	fmt.Println()
	pterm.Info.Println("Available plugins:")
	
	discoveryOpts := plugins.DefaultDiscoveryOptions()
	result, err := plugins.DiscoverAvailablePlugins(discoveryOpts)
	if err == nil {
		registry := plugins.GetBuiltinRegistry()
		for _, name := range result.BuiltinPlugins {
			pterm.Printf("  - %s (built-in)\n", name)
		}
		for _, manifest := range result.ExternalPlugins {
			pterm.Printf("  - %s (external)\n", manifest.Name)
		}
	}

	return fmt.Errorf("plugin name required")
}

// pluginGenerate implements the 'wails plugin generate' command
func pluginGenerate(f *struct{}) error {
	// Create logger
	logger := clilogger.New(os.Stdout)

	// Print banner
	app.PrintBanner()

	pterm.DefaultSection.Println("Plugin Code Generation")

	// Require project context
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	projectData, err := project.Load(cwd)
	if err != nil {
		return fmt.Errorf("no Wails project found in current directory. Please run this command from a Wails project root")
	}

	pterm.Info.Printf("Project: %s\n", projectData.Name)
	pterm.Info.Printf("Path: %s\n", projectData.Path)

	// TODO: Load plugin configuration from wails.json
	// For now, create basic config
	pluginConfig := &plugins.PluginConfig{
		ProjectRoot: projectData.Path,
		ProjectName: projectData.Name,
		Logger:      logger,
		Config:      make(map[string]interface{}),
		Environment: "development",
		Debug:       true,
		StartTime:   time.Now(),
	}

	// Create plugin manager
	managerConfig := &plugins.ManagerConfig{
		Logger:      logger,
		ProjectRoot: projectData.Path,
		ProjectName: projectData.Name,
	}
	manager := plugins.NewManager(managerConfig)

	// Initialize plugin manager
	ctx := context.Background()
	if err := manager.Initialize(ctx); err != nil {
		return fmt.Errorf("plugin manager initialization failed: %w", err)
	}
	defer manager.Shutdown(ctx)

	// Create generation context
	genCtx := &plugins.GenerationContext{
		Context:     ctx,
		ProjectRoot: projectData.Path,
		ProjectName: projectData.Name,
		OutputDir:   projectData.Path,
		Language:    "typescript", // Default to TypeScript
		Framework:   "react",      // Default framework
		Logger:      logger,
		PluginData:  make(map[string]interface{}),
	}

	// Execute code generation hooks
	pterm.Info.Println("Running code generation...")
	
	if err := manager.ExecuteHook("GenerateCode", genCtx); err != nil {
		return fmt.Errorf("code generation failed: %w", err)
	}

	pterm.Success.Println("Code generation completed successfully!")
	
	// TODO: Show generated files summary
	fmt.Println()
	pterm.Info.Println("Generated files will be written to the project directory based on plugin configuration.")

	return nil
}

// Helper functions for displaying plugin information

func displayPluginInfo(plugin plugins.Plugin, source string) {
	fmt.Println()
	
	// Basic information
	tableData := pterm.TableData{
		{"Name", plugin.Name()},
		{"Version", plugin.Version()},
		{"Description", plugin.Description()},
		{"Author", plugin.Author()},
		{"Source", source},
	}

	err := pterm.DefaultTable.WithHasHeader(false).WithData(tableData).Render()
	if err != nil {
		pterm.Error.Printf("Failed to display plugin info: %v\n", err)
		return
	}

	// Show capabilities
	fmt.Println()
	pterm.DefaultHeader.Println("Capabilities:")
	
	capabilities := getPluginCapabilities(plugin)
	if len(capabilities) > 0 {
		for _, capability := range capabilities {
			pterm.Printf("  ✓ %s\n", capability)
		}
	} else {
		pterm.Info.Println("  No special capabilities detected")
	}

	// Configuration example
	fmt.Println()
	pterm.DefaultHeader.Println("Configuration Example:")
	configExample := getConfigExample(plugin.Name())
	if configExample != "" {
		fmt.Println(configExample)
	} else {
		pterm.Info.Println("  No configuration required")
	}
}

func displayExternalPluginInfo(manifest *plugins.PluginManifest) {
	fmt.Println()
	
	// Basic information
	tableData := pterm.TableData{
		{"Name", manifest.Name},
		{"Version", manifest.Version},
		{"Description", manifest.Description},
		{"Author", manifest.Author},
		{"Source", "External"},
		{"Entry Point", manifest.EntryPoint},
		{"Type", manifest.Type},
	}

	err := pterm.DefaultTable.WithHasHeader(false).WithData(tableData).Render()
	if err != nil {
		pterm.Error.Printf("Failed to display plugin info: %v\n", err)
		return
	}

	// Show dependencies
	if manifest.Dependencies != nil && len(manifest.Dependencies) > 0 {
		fmt.Println()
		pterm.DefaultHeader.Println("Dependencies:")
		for dep, version := range manifest.Dependencies {
			pterm.Printf("  - %s: %s\n", dep, version)
		}
	}

	// Show configuration if available
	if len(manifest.Config) > 0 {
		fmt.Println()
		pterm.DefaultHeader.Println("Configuration Schema:")
		
		var configData interface{}
		if err := json.Unmarshal(manifest.Config, &configData); err == nil {
			if configJSON, err := json.MarshalIndent(configData, "  ", "  "); err == nil {
				fmt.Printf("  %s\n", string(configJSON))
			}
		}
	}
}

func getPluginCapabilities(plugin plugins.Plugin) []string {
	var capabilities []string
	
	// Check plugin interfaces using type assertions
	if _, ok := plugin.(plugins.CodeGenerator); ok {
		capabilities = append(capabilities, "Code Generation")
	}
	if _, ok := plugin.(plugins.BuildHook); ok {
		capabilities = append(capabilities, "Build Integration")
	}
	if _, ok := plugin.(plugins.DevServerExtension); ok {
		capabilities = append(capabilities, "Development Server Extension")
	}
	if _, ok := plugin.(plugins.FileWatcher); ok {
		capabilities = append(capabilities, "File Watching")
	}
	if _, ok := plugin.(plugins.TemplateProvider); ok {
		capabilities = append(capabilities, "Project Templates")
	}
	if _, ok := plugin.(plugins.AssetProcessor); ok {
		capabilities = append(capabilities, "Asset Processing")
	}
	
	return capabilities
}

func getConfigExample(pluginName string) string {
	// Provide configuration examples for known plugins
	examples := map[string]string{
		"basic-react": `{
  "plugins": {
    "basic-react": {
      "enabled": true,
      "config": {
        "typescript": true,
        "hooks": ["useState", "useEffect"]
      }
    }
  }
}`,
		"basic-vue": `{
  "plugins": {
    "basic-vue": {
      "enabled": true,
      "config": {
        "composition": true,
        "router": true
      }
    }
  }
}`,
		"dev-tools": `{
  "plugins": {
    "dev-tools": {
      "enabled": true,
      "config": {
        "hotReload": true,
        "debugMode": true
      }
    }
  }
}`,
		"wrf": `{
  "plugins": {
    "wrf": {
      "enabled": true,
      "config": {
        "queryPrefix": "use",
        "mutationPrefix": "use",
        "cacheTime": "5m"
      }
    }
  }
}`,
	}
	
	if example, exists := examples[pluginName]; exists {
		return example
	}
	
	return ""
}

// Register plugin commands with the CLI
func registerPluginCommands(app *clir.Cli) {
	// Create plugin command group
	plugin := app.NewSubCommand("plugin", "Plugin management commands")
	
	// Add description
	plugin.LongDescription(`Manage Wails plugins for enhanced development experience.

Plugins extend Wails with:
  - Code generation from Go to TypeScript
  - Framework-specific integrations  
  - Development tools and debugging
  - Custom build steps`)

	// Add subcommands
	plugin.NewSubCommandFunction("list", "List available plugins", pluginList)
	plugin.NewSubCommandFunction("info", "Show detailed information about a plugin", pluginInfo)
	plugin.NewSubCommandFunction("generate", "Run code generation for enabled plugins", pluginGenerate)
}