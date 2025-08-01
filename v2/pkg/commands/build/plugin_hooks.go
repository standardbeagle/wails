package build

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2/internal/project"
	"github.com/wailsapp/wails/v2/pkg/plugins"
)

// executePreBuildHooks runs pre-build plugins
func executePreBuildHooks(manager *plugins.Manager, ctx *plugins.BuildContext) error {
	hooks := manager.GetPluginsByType("BuildHook")
	if len(hooks) == 0 {
		return nil
	}

	for _, plugin := range hooks {
		hook := plugin.(plugins.BuildHook)

		ctx.Logger.Debug("Running pre-build hook", "plugin", plugin.Name())

		if err := hook.PreBuild(ctx); err != nil {
			return fmt.Errorf("plugin %s pre-build failed: %w", plugin.Name(), err)
		}
	}

	return nil
}

// executeCodeGeneration runs code generation plugins
func executeCodeGeneration(manager *plugins.Manager, project *project.Project, buildCtx *plugins.BuildContext) error {
	generators := manager.GetPluginsByType("CodeGenerator")
	if len(generators) == 0 {
		return nil
	}

	genCtx := &plugins.GenerationContext{
		Context:       buildCtx.Context,
		ProjectRoot:   project.Path,
		OutputDir:     project.Path,
		ProjectConfig: extractProjectConfig(project),
		Logger:        buildCtx.Logger,
	}

	for _, plugin := range generators {
		gen := plugin.(plugins.CodeGenerator)

		buildCtx.Logger.Debug("Running code generation", "plugin", plugin.Name())

		files, err := gen.GenerateCode(genCtx)
		if err != nil {
			return fmt.Errorf("plugin %s generation failed: %w", plugin.Name(), err)
		}

		// Write generated files
		for _, file := range files {
			if err := writeGeneratedFile(project.Path, file); err != nil {
				return fmt.Errorf("failed to write generated file %s: %w", file.Path, err)
			}
		}

		buildCtx.Logger.Info("Generated files", "plugin", plugin.Name(), "count", len(files))
	}

	return nil
}

// executePostBuildHooks runs post-build plugins
func executePostBuildHooks(manager *plugins.Manager, ctx *plugins.BuildContext) error {
	hooks := manager.GetPluginsByType("BuildHook")
	if len(hooks) == 0 {
		return nil
	}

	var errors []error

	for _, plugin := range hooks {
		hook := plugin.(plugins.BuildHook)

		ctx.Logger.Debug("Running post-build hook", "plugin", plugin.Name())

		if err := hook.PostBuild(ctx); err != nil {
			errors = append(errors, fmt.Errorf("plugin %s: %w", plugin.Name(), err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("post-build errors: %v", errors)
	}

	return nil
}

// writeGeneratedFile safely writes a generated file
func writeGeneratedFile(projectRoot string, file *plugins.GeneratedFile) error {
	fullPath := filepath.Join(projectRoot, file.Path)

	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Check if file exists and overwrite flag
	if !file.Overwrite {
		if _, err := os.Stat(fullPath); err == nil {
			// File exists, skip
			return nil
		}
	}

	// Write file with proper permissions
	mode := os.FileMode(file.Mode)
	if mode == 0 {
		mode = 0644
	}

	return os.WriteFile(fullPath, file.Content, mode)
}

// initializePluginManager creates and initializes the plugin manager
func initializePluginManager(config *project.Project) (*plugins.Manager, error) {
	// Check if plugins are configured
	if config.Plugins == nil || !config.Plugins.Enabled {
		return nil, nil
	}

	// Create manager with configuration
	manager := plugins.NewManager()

	// TODO: Initialize manager with project plugins
	// This will be completed once the options integration is available

	return manager, nil
}

// extractProjectConfig converts project data to plugin-friendly config
func extractProjectConfig(project *project.Project) map[string]interface{} {
	config := make(map[string]interface{})

	config["name"] = project.Name
	config["outputfilename"] = project.OutputFilename
	config["frontend:dir"] = project.GetFrontendDir()
	config["wailsjsdir"] = project.GetWailsJSDir()

	// Add build configuration
	if project.BuildTags != "" {
		config["build:tags"] = project.BuildTags
	}
	
	config["build:dir"] = project.GetBuildDir()

	return config
}

// getBuildMode converts build options to plugin build mode
func getBuildMode(options *Options) string {
	switch options.Mode {
	case Dev:
		return "development"
	case Production:
		return "production"
	case Debug:
		return "debug"
	default:
		return "unknown"
	}
}