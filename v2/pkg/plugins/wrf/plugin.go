package wrf

import (
	"context"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/plugins"
)

// WRFPlugin provides advanced React framework integration
type WRFPlugin struct {
	config      *Config
	logger      plugins.Logger
	projectRoot string
	parser      *Parser
	generator   *Generator
	devServer   *DevServer
}

// NewWRFPlugin creates the WRF plugin instance
func NewWRFPlugin() plugins.Plugin {
	return &WRFPlugin{}
}

// Plugin interface implementation
func (p *WRFPlugin) Name() string        { return "wrf" }
func (p *WRFPlugin) Version() string     { return "2.0.0" }
func (p *WRFPlugin) Description() string { 
	return "Wails React Framework - Advanced React integration with type-safe hooks"
}
func (p *WRFPlugin) Author() string      { return "WRF Team" }

// Initialize with enhanced configuration
func (p *WRFPlugin) Initialize(ctx context.Context, config *plugins.PluginConfig) error {
	p.projectRoot = config.ProjectRoot
	p.logger = config.Logger

	// Parse configuration
	p.config = DefaultConfig()
	if err := p.config.ParseFrom(config.Config); err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Initialize components
	p.parser = NewParser(p.config, p.logger)
	p.generator = NewGenerator(p.config, p.logger)
	p.devServer = NewDevServer(p.config, p.logger)

	p.logger.Info("WRF plugin initialized", 
		"version", p.Version(),
		"reactQuery", p.config.Generation.Hooks.Framework)

	return nil
}

// Shutdown the plugin
func (p *WRFPlugin) Shutdown(ctx context.Context) error {
	p.logger.Info("WRF plugin shutdown")
	return nil
}

// Health check
func (p *WRFPlugin) Health() error {
	if p.config == nil {
		return fmt.Errorf("plugin not initialized")
	}
	return nil
}

// GenerateCode with advanced features
func (p *WRFPlugin) GenerateCode(ctx *plugins.GenerationContext) ([]*plugins.GeneratedFile, error) {
	p.logger.Info("Starting WRF code generation")

	// Parse Go services with annotations
	services, err := p.parser.ParseServices(ctx.ProjectRoot)
	if err != nil {
		return nil, fmt.Errorf("service parsing failed: %w", err)
	}

	// Parse types for generation
	types, err := p.parser.ParseTypes(ctx.ProjectRoot)
	if err != nil {
		return nil, fmt.Errorf("type parsing failed: %w", err)
	}

	var files []*plugins.GeneratedFile

	// Generate TypeScript types with validation
	typeFiles, err := p.generator.GenerateTypes(types)
	if err != nil {
		return nil, err
	}
	files = append(files, typeFiles...)

	// Generate React Query hooks
	hookFiles, err := p.generator.GenerateHooks(services)
	if err != nil {
		return nil, err
	}
	files = append(files, hookFiles...)

	// Generate Zod schemas if enabled
	if p.config.Generation.Types.Validation == "zod" {
		schemaFiles, err := p.generator.GenerateSchemas(types)
		if err != nil {
			return nil, err
		}
		files = append(files, schemaFiles...)
	}

	// Generate event types
	eventFiles, err := p.generator.GenerateEventTypes(services)
	if err != nil {
		return nil, err
	}
	files = append(files, eventFiles...)

	// Generate index files
	files = append(files, p.generator.GenerateIndexFiles(services, types)...)

	p.logger.Info("WRF generation completed", "files", len(files))
	return files, nil
}

// SupportedLanguages returns supported frontend languages
func (p *WRFPlugin) SupportedLanguages() []string {
	return []string{"typescript", "javascript"}
}

// ConfigureDevServer adds WRF development tools
func (p *WRFPlugin) ConfigureDevServer(config plugins.DevServerConfigurator) error {
	return p.devServer.Configure(config)
}