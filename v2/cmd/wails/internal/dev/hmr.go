package dev

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HMRMessage represents an HMR protocol message
type HMRMessage struct {
	Type      string                 `json:"type"`
	Timestamp int64                  `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// HMR message types
const (
	HMRTypeReload      = "reload"
	HMRTypeUpdate      = "update"
	HMRTypeGenerated   = "generated"
	HMRTypeError       = "error"
	HMRTypeConnected   = "connected"
	HMRTypeWRFRegenerated = "wrf-regenerated"
	HMRTypeWRFError    = "wrf-error"
)

// DevServer represents the development server
type DevServer struct {
	projectRoot string
	wsHub       *WebSocketHub
	logger      interface{}
}

// WebSocketHub manages WebSocket connections
type WebSocketHub struct {
	broadcast chan []byte
	clients   map[*WebSocketClient]bool
}

// WebSocketClient represents a WebSocket client connection
type WebSocketClient struct {
	conn interface{}
}

// sendHMRUpdate sends HMR update to connected clients
func (s *DevServer) sendHMRUpdate(updateType, filePath string) {
	relPath, _ := filepath.Rel(s.projectRoot, filePath)

	msg := HMRMessage{
		Type:      HMRTypeUpdate,
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"updateType": updateType,
			"file":       relPath,
			"preserve":   shouldPreserveState(filePath),
		},
	}

	// Special handling for generated files
	if updateType == "generated" {
		msg.Type = HMRTypeGenerated
		msg.Data["generated"] = true
		msg.Data["hooks"] = extractHooksFromFile(filePath)
	}

	s.broadcastHMR(msg)
}

// broadcastHMR sends message to all connected clients
func (s *DevServer) broadcastHMR(msg HMRMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		// Log error (simplified for now)
		fmt.Printf("Failed to marshal HMR message: %v\n", err)
		return
	}

	if s.wsHub != nil {
		s.wsHub.broadcast <- data
	}
}

// BroadcastJSON sends JSON message to all clients
func (s *DevServer) BroadcastJSON(data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Failed to marshal JSON: %v\n", err)
		return
	}

	if s.wsHub != nil {
		s.wsHub.broadcast <- jsonData
	}
}

// triggerRebuild triggers a rebuild based on file type
func (s *DevServer) triggerRebuild(buildType, filePath string) {
	msg := HMRMessage{
		Type:      HMRTypeReload,
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"buildType": buildType,
			"file":      filePath,
			"reason":    fmt.Sprintf("%s file changed", buildType),
		},
	}

	s.broadcastHMR(msg)
}

// shouldPreserveState determines if React state should be preserved
func shouldPreserveState(filePath string) bool {
	// Preserve state for:
	// - Style files
	// - Generated hook files
	// - Type definition files

	ext := filepath.Ext(filePath)
	base := filepath.Base(filePath)

	switch ext {
	case ".css", ".scss", ".less":
		return true
	case ".d.ts":
		return true
	}

	// Generated files preserve state
	if strings.Contains(filePath, "/generated/") {
		return true
	}

	// Hook files preserve state
	if strings.HasSuffix(base, "Hooks.ts") || strings.HasSuffix(base, "Hooks.tsx") {
		return true
	}

	return false
}

// extractHooksFromFile extracts hook names from generated file
func extractHooksFromFile(filePath string) []string {
	// Simple extraction - in real implementation, parse the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

	var hooks []string
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "export function use") {
			// Extract hook name
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				hookName := strings.TrimSuffix(parts[2], "(")
				hooks = append(hooks, hookName)
			}
		}
	}

	return hooks
}

// HMR client code that gets injected into frontend
const hmrClientCode = `
// WRF HMR Client
(function() {
  let ws;
  let reconnectInterval = 1000;
  let maxReconnectInterval = 30000;
  let reconnectDecay = 1.5;
  let timeoutInterval = 2000;

  function connect() {
    ws = new WebSocket('ws://localhost:34115/ws');
    
    ws.onopen = function() {
      console.log('[HMR] Connected');
      reconnectInterval = 1000;
    };
    
    ws.onmessage = function(event) {
      const msg = JSON.parse(event.data);
      
      switch (msg.type) {
        case 'wrf-regenerated':
          handleWRFRegeneration(msg.data);
          break;
        case 'wrf-error':
          handleWRFError(msg.data);
          break;
        case 'update':
          handleUpdate(msg.data);
          break;
        case 'reload':
          handleReload(msg.data);
          break;
        case 'error':
          handleError(msg.data);
          break;
        default:
          console.log('[HMR] Unknown message type:', msg.type);
          break;
      }
    };
    
    ws.onclose = function() {
      console.log('[HMR] Disconnected');
      setTimeout(function() {
        reconnectInterval = Math.min(maxReconnectInterval, reconnectInterval * reconnectDecay);
        connect();
      }, reconnectInterval);
    };
    
    ws.onerror = function(err) {
      console.error('[HMR] Connection error:', err);
    };
  }
  
  function handleWRFRegeneration(data) {
    console.log('[WRF] Code regenerated', data);
    
    // Preserve React Query cache
    if (window.__REACT_QUERY_CLIENT__) {
      const queryClient = window.__REACT_QUERY_CLIENT__;
      
      // Update only affected queries
      data.hooks.forEach(hookFile => {
        // Invalidate queries related to this hook
        const hookName = extractHookName(hookFile);
        queryClient.invalidateQueries({ 
          predicate: (query) => query.queryKey[0] === hookName 
        });
      });
    }
    
    // Hot reload the modules
    if (module.hot) {
      data.hooks.forEach(file => {
        module.hot.accept('./' + file);
      });
      
      data.types.forEach(file => {
        module.hot.accept('./' + file);
      });
    } else {
      // Fallback to page reload if HMR not available
      window.location.reload();
    }
  }
  
  function handleWRFError(data) {
    console.error('[WRF] Generation error:', data.error);
    
    // Show error overlay
    showErrorOverlay({
      type: 'WRF Generation Error',
      message: data.error,
      plugin: data.plugin,
      timestamp: new Date(data.timestamp * 1000)
    });
  }
  
  function handleUpdate(data) {
    console.log('[HMR] File updated:', data.file);
    
    if (data.preserve && module.hot) {
      // Preserve state update
      module.hot.accept();
    } else {
      // Full reload
      window.location.reload();
    }
  }
  
  function handleReload(data) {
    console.log('[HMR] Reload requested:', data.reason);
    window.location.reload();
  }
  
  function handleError(data) {
    console.error('[HMR] Error:', data.error);
    showErrorOverlay(data);
  }
  
  function extractHookName(hookFile) {
    // Extract hook name from file path
    const fileName = hookFile.split('/').pop();
    return fileName.replace(/Hooks\.(ts|tsx)$/, '');
  }
  
  function showErrorOverlay(error) {
    // Create error overlay if it doesn't exist
    let overlay = document.getElementById('wrf-error-overlay');
    if (!overlay) {
      overlay = document.createElement('div');
      overlay.id = 'wrf-error-overlay';
      overlay.style.cssText = `
        position: fixed;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background: rgba(0, 0, 0, 0.8);
        color: white;
        font-family: monospace;
        font-size: 14px;
        padding: 20px;
        z-index: 9999;
        overflow: auto;
      `;
      
      overlay.onclick = function() {
        overlay.remove();
      };
      
      document.body.appendChild(overlay);
    }
    
    overlay.innerHTML = `
      <div style="max-width: 800px; margin: 0 auto;">
        <h2 style="color: #ff6b6b; margin-bottom: 20px;">${error.type || 'Error'}</h2>
        <pre style="background: #2d2d2d; padding: 15px; border-radius: 5px; overflow: auto;">
${error.message}
        </pre>
        <p style="margin-top: 20px; opacity: 0.7;">
          Click anywhere to dismiss
        </p>
      </div>
    `;
  }
  
  // Initialize connection
  connect();
  
  // Expose to global scope for debugging
  window.__WRF_HMR__ = { ws, reconnect: connect };
})();
`

// GetHMRClientCode returns the HMR client code
func GetHMRClientCode() string {
	return hmrClientCode
}