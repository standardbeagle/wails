package wrf

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/plugins"
)

// GetWatchPaths returns paths to watch for WRF plugin
func (p *WRFPlugin) GetWatchPaths(projectRoot string) ([]string, error) {
	var paths []string

	// Watch configured source paths
	for _, sourcePath := range p.config.Go.SourcePaths {
		absPath := filepath.Join(projectRoot, sourcePath)

		// Handle glob patterns
		if strings.Contains(sourcePath, "*") {
			matches, err := filepath.Glob(absPath)
			if err != nil {
				return nil, err
			}
			paths = append(paths, matches...)
		} else {
			paths = append(paths, absPath)
		}
	}

	// Watch WRF configuration
	paths = append(paths, filepath.Join(projectRoot, "wails.json"))

	return paths, nil
}

// OnFileChanged handles file changes for WRF plugin
func (p *WRFPlugin) OnFileChanged(ctx *plugins.FileChangeContext) error {
	// Only process Go files
	if !strings.HasSuffix(ctx.FilePath, ".go") {
		return nil
	}

	// Check if file is in watched paths
	if !p.isWatchedFile(ctx.FilePath) {
		return nil
	}

	p.logger.Info("WRF detected Go file change", "file", ctx.FilePath)

	// Mark for regeneration
	p.markForRegeneration(ctx.FilePath)

	// If in watch mode, trigger regeneration
	if p.config.Go.WatchMode {
		// Debounced regeneration
		p.scheduleRegeneration(ctx)
	}

	return nil
}

// isWatchedFile checks if a file is in the watched paths
func (p *WRFPlugin) isWatchedFile(filePath string) bool {
	for _, sourcePath := range p.config.Go.SourcePaths {
		absSourcePath := filepath.Join(p.projectRoot, sourcePath)

		// Handle glob patterns
		if strings.Contains(sourcePath, "*") {
			matches, err := filepath.Glob(absSourcePath)
			if err != nil {
				continue
			}

			for _, match := range matches {
				if strings.HasPrefix(filePath, match) {
					return true
				}
			}
		} else {
			if strings.HasPrefix(filePath, absSourcePath) {
				return true
			}
		}
	}

	return false
}

// markForRegeneration marks a file for regeneration
func (p *WRFPlugin) markForRegeneration(filePath string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.pendingRegeneration == nil {
		p.pendingRegeneration = make(map[string]time.Time)
	}

	p.pendingRegeneration[filePath] = time.Now()
}

// scheduleRegeneration schedules code regeneration with debouncing
func (p *WRFPlugin) scheduleRegeneration(ctx *plugins.FileChangeContext) {
	if p.regenerateDebouncer == nil {
		p.regenerateDebouncer = NewRegenerateDebouncer(1 * time.Second)
	}

	p.regenerateDebouncer.Debounce("regenerate", func() {
		p.logger.Info("WRF triggering regeneration")

		genCtx := &plugins.GenerationContext{
			Context:       ctx.Context,
			ProjectRoot:   ctx.ProjectRoot,
			ProjectConfig: p.extractProjectConfig(),
			Logger:        ctx.Logger,
		}

		files, err := p.GenerateCode(genCtx)
		if err != nil {
			p.logger.Error("Regeneration failed", "error", err)

			// Send error to frontend
			p.sendHMRError(err)
			return
		}

		// Write generated files
		for _, file := range files {
			if err := p.writeGeneratedFile(file); err != nil {
				p.logger.Error("Failed to write file", "path", file.Path, "error", err)
			}
		}

		// Notify frontend of generated files
		p.sendGeneratedFilesUpdate(files)

		// Clear pending regeneration
		p.mu.Lock()
		p.pendingRegeneration = make(map[string]time.Time)
		p.mu.Unlock()
	})
}

// extractProjectConfig extracts project configuration for generation
func (p *WRFPlugin) extractProjectConfig() interface{} {
	// Return a simplified project config
	return map[string]interface{}{
		"projectRoot": p.projectRoot,
		"config":      p.config,
	}
}

// writeGeneratedFile writes a generated file to disk
func (p *WRFPlugin) writeGeneratedFile(file *plugins.GeneratedFile) error {
	// Ensure directory exists
	dir := filepath.Dir(file.Path)
	if err := p.ensureDir(dir); err != nil {
		return err
	}

	// Write file content
	return p.writeFile(file.Path, file.Content)
}

// RegenerateDebouncer handles debounced regeneration
type RegenerateDebouncer struct {
	delay  time.Duration
	timer  *time.Timer
	mu     sync.Mutex
}

// NewRegenerateDebouncer creates a new regenerate debouncer
func NewRegenerateDebouncer(delay time.Duration) *RegenerateDebouncer {
	return &RegenerateDebouncer{
		delay: delay,
	}
}

// Debounce executes the function after the delay
func (d *RegenerateDebouncer) Debounce(key string, fn func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Cancel existing timer
	if d.timer != nil {
		d.timer.Stop()
	}

	// Create new timer
	d.timer = time.AfterFunc(d.delay, fn)
}