package react

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/plugins"
)

// BasicReactPlugin provides basic React integration
type BasicReactPlugin struct {
	config      *Config
	logger      plugins.Logger
	projectRoot string
}

// NewBasicReactPlugin creates a new instance
func NewBasicReactPlugin() plugins.Plugin {
	return &BasicReactPlugin{}
}

// Plugin interface implementation
func (p *BasicReactPlugin) Name() string        { return "basic-react" }
func (p *BasicReactPlugin) Version() string     { return "1.0.0" }
func (p *BasicReactPlugin) Description() string { 
	return "Basic React support with TypeScript bindings and hooks"
}
func (p *BasicReactPlugin) Author() string      { return "Wails Team" }

// Initialize the plugin
func (p *BasicReactPlugin) Initialize(ctx context.Context, config *plugins.PluginConfig) error {
	p.projectRoot = config.ProjectRoot
	p.logger = config.Logger

	// Parse plugin configuration
	p.config = &Config{
		TypesOutput:     "frontend/src/types/generated.ts",
		HooksOutput:     "frontend/src/hooks/generated",
		IncludeComments: true,
	}

	if err := p.config.ParseFrom(config.Config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	p.logger.Info("Basic React plugin initialized", 
		"typesOutput", p.config.TypesOutput,
		"hooksOutput", p.config.HooksOutput)

	return nil
}

// Shutdown the plugin
func (p *BasicReactPlugin) Shutdown(ctx context.Context) error {
	p.logger.Info("Basic React plugin shutdown")
	return nil
}

// Health check
func (p *BasicReactPlugin) Health() error {
	if p.config == nil {
		return fmt.Errorf("plugin not initialized")
	}
	return nil
}

// GenerateCode generates TypeScript types and React hooks
func (p *BasicReactPlugin) GenerateCode(ctx *plugins.GenerationContext) ([]*plugins.GeneratedFile, error) {
	p.logger.Info("Starting code generation")

	// Parse Go source files
	services, err := p.parseGoServices(ctx.ProjectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to parse services: %w", err)
	}

	if len(services) == 0 {
		p.logger.Info("No services found for generation")
		return nil, nil
	}

	var files []*plugins.GeneratedFile

	// Generate TypeScript types
	typesFile, err := p.generateTypes(services)
	if err != nil {
		return nil, fmt.Errorf("failed to generate types: %w", err)
	}
	files = append(files, typesFile)

	// Generate React hooks
	hookFiles, err := p.generateHooks(services)
	if err != nil {
		return nil, fmt.Errorf("failed to generate hooks: %w", err)
	}
	files = append(files, hookFiles...)

	// Generate index file
	indexFile := p.generateIndex(services)
	files = append(files, indexFile)

	p.logger.Info("Code generation completed", "files", len(files))
	return files, nil
}

// SupportedLanguages returns supported frontend languages
func (p *BasicReactPlugin) SupportedLanguages() []string {
	return []string{"typescript", "javascript"}
}