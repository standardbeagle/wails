package wrf

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/plugins"
)

// sendGeneratedFilesUpdate notifies frontend of regenerated files
func (p *WRFPlugin) sendGeneratedFilesUpdate(files []*plugins.GeneratedFile) {
	var hooks []string
	var types []string
	var schemas []string

	// Categorize generated files
	for _, file := range files {
		relPath, _ := filepath.Rel(p.projectRoot, file.Path)

		if strings.Contains(relPath, "/hooks/") {
			hooks = append(hooks, relPath)
		} else if strings.Contains(relPath, "/types/") {
			types = append(types, relPath)
		} else if strings.Contains(relPath, "/schemas/") {
			schemas = append(schemas, relPath)
		}
	}

	// Send specialized HMR message
	msg := map[string]interface{}{
		"type": "wrf-regenerated",
		"data": map[string]interface{}{
			"hooks":        hooks,
			"types":        types,
			"schemas":      schemas,
			"preserveState": true,
			"timestamp":    time.Now().Unix(),
		},
	}

	if p.devServer != nil {
		p.devServer.BroadcastJSON(msg)
	}
}

// sendHMRError sends error message to frontend
func (p *WRFPlugin) sendHMRError(err error) {
	msg := map[string]interface{}{
		"type": "wrf-error",
		"data": map[string]interface{}{
			"error":     err.Error(),
			"plugin":    "wrf",
			"action":    "regeneration",
			"timestamp": time.Now().Unix(),
		},
	}

	if p.devServer != nil {
		p.devServer.BroadcastJSON(msg)
	}
}

// HMRDevServer interface for dev server communication
type HMRDevServer interface {
	BroadcastJSON(data interface{})
}

// SetDevServer sets the development server for HMR communication
func (p *WRFPlugin) SetDevServer(server HMRDevServer) {
	p.devServer = server
}

// HMRState tracks HMR-related state
type HMRState struct {
	lastRegeneration  time.Time
	affectedFiles     map[string]time.Time
	preserveState     bool
	mu                sync.RWMutex
}

// NewHMRState creates a new HMR state tracker
func NewHMRState() *HMRState {
	return &HMRState{
		affectedFiles: make(map[string]time.Time),
		preserveState: true,
	}
}

// TrackFileChange tracks a file change for HMR
func (h *HMRState) TrackFileChange(filePath string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.affectedFiles[filePath] = time.Now()
}

// GetAffectedFiles returns files affected since last regeneration
func (h *HMRState) GetAffectedFiles() map[string]time.Time {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make(map[string]time.Time)
	for file, timestamp := range h.affectedFiles {
		if timestamp.After(h.lastRegeneration) {
			result[file] = timestamp
		}
	}

	return result
}

// MarkRegeneration marks a regeneration as complete
func (h *HMRState) MarkRegeneration() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.lastRegeneration = time.Now()
}

// ShouldPreserveState returns whether React state should be preserved
func (h *HMRState) ShouldPreserveState() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.preserveState
}

// SetPreserveState sets whether to preserve React state
func (h *HMRState) SetPreserveState(preserve bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.preserveState = preserve
}

// GenerateHMRInjection generates HMR injection code for the frontend
func (p *WRFPlugin) GenerateHMRInjection() string {
	return fmt.Sprintf(`
// WRF HMR Integration
(function() {
  if (typeof window === 'undefined') return;
  
  // Initialize WRF HMR state
  window.__WRF_HMR_STATE__ = {
    preserveCache: %t,
    lastUpdate: Date.now(),
    affectedFiles: new Set(),
    errorOverlay: null
  };
  
  // Register WRF-specific HMR handlers
  if (module.hot) {
    module.hot.accept();
    
    // Handle WRF regeneration
    module.hot.addStatusHandler(function(status) {
      if (status === 'apply') {
        console.log('[WRF HMR] Applying updates with state preservation');
      }
    });
  }
  
  // React Query integration
  if (window.QueryClient) {
    const queryClient = window.QueryClient;
    
    // Store original cache
    window.__WRF_HMR_STATE__.queryCache = queryClient.getQueryCache();
  }
})();
`, p.config.Development.HMR)
}

// GetHMRConfig returns HMR configuration for frontend
func (p *WRFPlugin) GetHMRConfig() map[string]interface{} {
	return map[string]interface{}{
		"enabled":       p.config.Development.HMR,
		"preserveState": true,
		"devtools":      p.config.Development.DevTools,
		"endpoints": map[string]string{
			"websocket": "ws://localhost:34115/ws",
			"inspector": "/wrf/inspector",
			"types":     "/wrf/types",
			"events":    "/wrf/events",
		},
	}
}