package wrf

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/wailsapp/wails/v2/pkg/plugins"
)

// DevServer provides development tools for WRF
type DevServer struct {
	config *Config
	logger plugins.Logger
	
	// Connected clients for event streaming
	clients map[string]bool
}

// NewDevServer creates a new dev server instance
func NewDevServer(config *Config, logger plugins.Logger) *DevServer {
	return &DevServer{
		config:  config,
		logger:  logger,
		clients: make(map[string]bool),
	}
}

// Configure adds WRF development tools to the dev server
func (ds *DevServer) Configure(config plugins.DevServerConfigurator) error {
	if !ds.config.Development.DevTools {
		return nil
	}

	// Add service inspector endpoint
	config.AddHandler("/wrf/inspector", http.HandlerFunc(ds.serveInspector))
	
	// Add type browser endpoint
	config.AddHandler("/wrf/types", http.HandlerFunc(ds.serveTypeBrowser))
	
	// Add event monitor endpoint (simplified for now)
	config.AddHandler("/wrf/events", http.HandlerFunc(ds.serveEvents))
	
	// Add React Query devtools bridge
	config.AddHandler("/wrf/devtools", http.HandlerFunc(ds.serveDevTools))
	
	// Add annotation viewer
	config.AddHandler("/wrf/annotations", http.HandlerFunc(ds.serveAnnotations))
	
	// Add code generation status
	config.AddHandler("/wrf/generation", http.HandlerFunc(ds.serveGenerationStatus))

	ds.logger.Info("WRF development tools configured")
	return nil
}

// serveInspector provides service inspection UI
func (ds *DevServer) serveInspector(w http.ResponseWriter, r *http.Request) {
	// This would integrate with the parser to show current services
	// For now, return a simple JSON response
	
	response := map[string]interface{}{
		"services": []map[string]interface{}{
			{
				"name": "UserService",
				"methods": []map[string]interface{}{
					{
						"name": "GetUser",
						"type": "query",
						"annotations": map[string]string{
							"cache": "5m",
							"events": "user:updated,user:deleted",
						},
					},
					{
						"name": "UpdateUser", 
						"type": "mutation",
						"annotations": map[string]string{
							"optimistic": "true",
							"invalidate": "UserService:GetUser",
						},
					},
				},
			},
		},
		"stats": map[string]interface{}{
			"totalServices": 1,
			"totalMethods": 2,
			"lastGenerated": time.Now().Format(time.RFC3339),
			"framework": ds.config.Generation.Hooks.Framework,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// serveTypeBrowser provides type information browser
func (ds *DevServer) serveTypeBrowser(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"types": []map[string]interface{}{
			{
				"name": "User",
				"fields": []map[string]interface{}{
					{
						"name": "id",
						"type": "number",
						"optional": false,
					},
					{
						"name": "name",
						"type": "string", 
						"optional": false,
					},
					{
						"name": "email",
						"type": "string",
						"optional": true,
					},
				},
			},
		},
		"generation": map[string]interface{}{
			"validation": ds.config.Generation.Types.Validation,
			"schemas": ds.config.Generation.Types.IncludeSchemas,
			"nullable": ds.config.Generation.Types.NullableFields,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// serveEvents provides event monitoring information
func (ds *DevServer) serveEvents(w http.ResponseWriter, r *http.Request) {
	if !ds.config.Development.EventDebugging {
		http.Error(w, "Event debugging disabled", http.StatusForbidden)
		return
	}

	response := map[string]interface{}{
		"events": []map[string]interface{}{
			{
				"name": "user:updated",
				"type": "mutation",
				"lastSeen": time.Now().Unix(),
			},
			{
				"name": "user:created",
				"type": "mutation", 
				"lastSeen": time.Now().Unix() - 300,
			},
		},
		"status": "monitoring",
		"clients": len(ds.clients),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// BroadcastEvent logs an event (simplified without WebSocket)
func (ds *DevServer) BroadcastEvent(eventName string, data interface{}) {
	if !ds.config.Development.EventDebugging {
		return
	}

	ds.logger.Info("Event broadcast", "event", eventName, "data", data)
}

// serveDevTools provides React Query devtools integration
func (ds *DevServer) serveDevTools(w http.ResponseWriter, r *http.Request) {
	if ds.config.Generation.Hooks.Framework != "react-query" {
		http.Error(w, "React Query devtools only available with react-query framework", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"devtools": map[string]interface{}{
			"enabled": true,
			"version": "4.0.0",
			"framework": "react-query",
		},
		"queries": []map[string]interface{}{
			{
				"queryKey": []string{"UserService", "getUser", "1"},
				"status": "success",
				"lastUpdated": time.Now().Unix(),
				"data": map[string]interface{}{
					"id": 1,
					"name": "John Doe",
					"email": "john@example.com",
				},
			},
		},
		"mutations": []map[string]interface{}{
			{
				"mutationKey": []string{"UserService", "updateUser"},
				"status": "idle",
				"lastUpdated": time.Now().Unix(),
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// serveAnnotations provides annotation information
func (ds *DevServer) serveAnnotations(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"annotations": map[string]interface{}{
			"query": map[string]interface{}{
				"description": "Marks a method as a query operation with caching",
				"parameters": []map[string]interface{}{
					{"name": "cache", "type": "duration", "description": "Cache duration"},
					{"name": "events", "type": "[]string", "description": "Invalidation events"},
					{"name": "staleTime", "type": "duration", "description": "Stale time"},
					{"name": "enabled", "type": "bool", "description": "Query enabled state"},
				},
			},
			"mutation": map[string]interface{}{
				"description": "Marks a method as a mutation operation",
				"parameters": []map[string]interface{}{
					{"name": "optimistic", "type": "bool", "description": "Enable optimistic updates"},
					{"name": "invalidate", "type": "[]string", "description": "Queries to invalidate"},
					{"name": "events", "type": "[]string", "description": "Events to emit"},
					{"name": "retryCount", "type": "int", "description": "Retry attempts"},
				},
			},
			"subscription": map[string]interface{}{
				"description": "Marks a method as a subscription operation",
				"parameters": []map[string]interface{}{
					{"name": "eventName", "type": "string", "description": "Event to subscribe to"},
					{"name": "autoStart", "type": "bool", "description": "Auto-start subscription"},
					{"name": "buffer", "type": "int", "description": "Event buffer size"},
				},
			},
			"event": map[string]interface{}{
				"description": "Marks a method as an event emitter",
				"parameters": []map[string]interface{}{
					{"name": "eventName", "type": "string", "description": "Event name"},
					{"name": "global", "type": "bool", "description": "Global event"},
					{"name": "typed", "type": "bool", "description": "Type-safe event"},
				},
			},
		},
		"examples": map[string]interface{}{
			"query": "@wrf:query(cache: \"5m\", events: [\"user:updated\"])",
			"mutation": "@wrf:mutation(optimistic: true, invalidate: [\"UserService:GetUser\"])",
			"subscription": "@wrf:subscription(eventName: \"user:status\", autoStart: true)",
			"event": "@wrf:event(eventName: \"user:created\", typed: true)",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// serveGenerationStatus provides code generation status
func (ds *DevServer) serveGenerationStatus(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status": "ready",
		"lastGeneration": time.Now().Format(time.RFC3339),
		"config": map[string]interface{}{
			"framework": ds.config.Generation.Hooks.Framework,
			"validation": ds.config.Generation.Types.Validation,
			"hmr": ds.config.Development.HMR,
		},
		"outputs": map[string]interface{}{
			"hooks": ds.config.Generation.Hooks.Output,
			"types": ds.config.Generation.Types.Output,
			"events": ds.config.Generation.Events.Output,
		},
		"stats": map[string]interface{}{
			"services": 1,
			"methods": 2,
			"types": 1,
			"fields": 3,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}