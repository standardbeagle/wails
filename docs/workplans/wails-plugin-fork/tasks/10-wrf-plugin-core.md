# Task: WRF Plugin Core Implementation
**Generated from Master Planning**: 2024-01-15
**Context Package**: `/requests/wails-plugin-fork/context/`

## Task Sizing Assessment
**File Count**: 8-10 files
**Estimated Time**: 90 minutes
**Token Estimate**: 150k tokens
**Complexity Level**: 4 (Atomic Large)
**Parallelization Benefit**: MEDIUM - Can be split but core is complex
**Atomicity Assessment**: ✅ ATOMIC - Core WRF functionality
**Boundary Analysis**: ✅ CLEAR - Separate plugin package

## Persona Assignment
**Persona**: Software Engineer
**Expertise Required**: Go, TypeScript, React, AST parsing, code generation
**Worktree**: `~/work/wrf/wails-plugin-fork/`

## Context Summary
**Risk Level**: HIGH - Complex plugin with advanced features
**Integration Points**: All plugin interfaces, React Query, TypeScript
**Architecture Pattern**: Advanced code generation with annotations
**Similar Reference**: Basic React plugin as foundation

### Task Scope Boundaries
**MODIFY Zone** (Direct Changes):
```yaml
primary_files:
  - /v2/pkg/plugins/wrf/plugin.go              # Main WRF plugin
  - /v2/pkg/plugins/wrf/parser.go              # Advanced Go parsing
  - /v2/pkg/plugins/wrf/annotations.go         # Annotation parsing
  - /v2/pkg/plugins/wrf/generator.go           # Code generation
  - /v2/pkg/plugins/wrf/templates.go           # Template system
  - /v2/pkg/plugins/wrf/react_query.go         # React Query integration
  - /v2/pkg/plugins/wrf/types.go               # Type definitions
  - /v2/pkg/plugins/wrf/config.go              # Configuration
  - /v2/pkg/plugins/wrf/dev_server.go          # Dev server extensions
  - /v2/pkg/plugins/wrf/wrf_test.go            # Tests
```

**REVIEW Zone** (Check for Impact):
```yaml
check_patterns:
  - /v2/pkg/plugins/builtin.go                # Plugin registration
  - /v2/pkg/plugins/builtin/react/           # Basic React patterns
```

**IGNORE Zone** (Do Not Touch):
```yaml
ignore_completely:
  - /v2/internal/                            # Core Wails internals
  - /v2/cmd/                                 # CLI already done
```

## Task Requirements
**Objective**: Implement WRF plugin with advanced React integration and TypeScript generation

**Success Criteria**:
- [ ] Parse Go annotations for query/mutation metadata
- [ ] Generate React Query hooks with caching
- [ ] Support optimistic updates and invalidation
- [ ] Generate Zod validation schemas
- [ ] Real-time event integration
- [ ] Dev server tools for debugging
- [ ] Comprehensive type safety

**Validation Commands**:
```bash
# Enable WRF plugin
echo '{"plugins":{"wrf":{"enabled":true}}}' > wails.json

# Generate code
wails plugin generate

# Check generated files
ls frontend/src/hooks/generated/
ls frontend/src/types/generated.ts
ls frontend/src/schemas/generated/

# Test dev tools
curl http://localhost:34115/plugin/wrf/inspector
```

## Implementation Details

### 1. Main WRF Plugin (`plugin.go`)
```go
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

// Additional interface implementations...
```

### 2. Annotation Parser (`annotations.go`)
```go
package wrf

import (
    "fmt"
    "regexp"
    "strings"
)

// Annotation represents a WRF annotation
type Annotation struct {
    Type   string            // query, mutation, subscription
    Params map[string]string // Parameters like cache, invalidate
}

// ParseAnnotations extracts WRF annotations from comments
func ParseAnnotations(comments []string) []Annotation {
    var annotations []Annotation
    
    // Pattern: @wrf:type(param: "value", param2: value)
    pattern := regexp.MustCompile(`@wrf:(\w+)(?:\((.*?)\))?`)
    
    for _, comment := range comments {
        matches := pattern.FindStringSubmatch(comment)
        if len(matches) < 2 {
            continue
        }
        
        ann := Annotation{
            Type:   matches[1],
            Params: make(map[string]string),
        }
        
        // Parse parameters if present
        if len(matches) > 2 && matches[2] != "" {
            params := parseParameters(matches[2])
            ann.Params = params
        }
        
        annotations = append(annotations, ann)
    }
    
    return annotations
}

// parseParameters parses annotation parameters
func parseParameters(paramStr string) map[string]string {
    params := make(map[string]string)
    
    // Simple parser for key: value pairs
    parts := strings.Split(paramStr, ",")
    for _, part := range parts {
        kv := strings.SplitN(strings.TrimSpace(part), ":", 2)
        if len(kv) == 2 {
            key := strings.TrimSpace(kv[0])
            value := strings.Trim(strings.TrimSpace(kv[1]), `"'`)
            params[key] = value
        }
    }
    
    return params
}

// Common annotation patterns
const (
    AnnQuery        = "query"
    AnnMutation     = "mutation"
    AnnSubscription = "subscription"
    AnnEvent        = "event"
)

// QueryAnnotation extracts query-specific parameters
type QueryAnnotation struct {
    Cache      string   // Cache duration
    Events     []string // Invalidation events
    StaleTime  string   // Stale time
    RefetchOn  []string // Refetch conditions
}

// MutationAnnotation extracts mutation-specific parameters
type MutationAnnotation struct {
    Optimistic  bool     // Enable optimistic updates
    Invalidate  []string // Query keys to invalidate
    Events      []string // Events to emit
    OnSuccess   string   // Success handler
}
```

### 3. React Query Generator (`react_query.go`)
```go
package wrf

import (
    "bytes"
    "fmt"
    "strings"
    "text/template"
)

// generateReactQueryHook generates a React Query hook
func (g *Generator) generateReactQueryHook(service Service, method Method) (string, error) {
    var buf bytes.Buffer
    
    // Determine hook type based on annotations
    hookType := g.determineHookType(method)
    
    switch hookType {
    case "query":
        return g.generateQueryHook(service, method)
    case "mutation":
        return g.generateMutationHook(service, method)
    case "subscription":
        return g.generateSubscriptionHook(service, method)
    default:
        return g.generateBasicHook(service, method)
    }
}

// generateQueryHook generates a React Query query hook
func (g *Generator) generateQueryHook(service Service, method Method) (string, error) {
    tmplStr := `
export function use{{.MethodName}}({{.Params}}) {
  return useQuery({
    queryKey: [{{.QueryKey}}],
    queryFn: async () => {
      const result = await window.go.{{.ServiceName}}.{{.MethodName}}({{.ParamNames}});
      return result;
    },
    {{- if .CacheTime}}
    cacheTime: {{.CacheTime}},
    {{- end}}
    {{- if .StaleTime}}
    staleTime: {{.StaleTime}},
    {{- end}}
    {{- if .RefetchInterval}}
    refetchInterval: {{.RefetchInterval}},
    {{- end}}
    {{- if .Events}}
    // Auto-invalidation on events
    onMount: (queryClient) => {
      const unsubscribes = [
        {{- range .Events}}
        EventsOn('{{.}}', () => {
          queryClient.invalidateQueries({ queryKey: [{{$.QueryKey}}] });
        }),
        {{- end}}
      ];
      
      return () => {
        unsubscribes.forEach(fn => fn());
      };
    },
    {{- end}}
  });
}
`
    
    tmpl := template.Must(template.New("query").Parse(tmplStr))
    
    // Extract query annotation
    queryAnn := extractQueryAnnotation(method.Annotations)
    
    data := map[string]interface{}{
        "MethodName":      method.Name,
        "ServiceName":     service.Name,
        "Params":          formatHookParams(method.Params),
        "ParamNames":      formatParamNames(method.Params),
        "QueryKey":        generateQueryKey(service.Name, method.Name, method.Params),
        "CacheTime":       parseDuration(queryAnn.Cache),
        "StaleTime":       parseDuration(queryAnn.StaleTime),
        "Events":          queryAnn.Events,
        "RefetchInterval": queryAnn.RefetchInterval,
    }
    
    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, data); err != nil {
        return "", err
    }
    
    return buf.String(), nil
}

// generateMutationHook generates a React Query mutation hook
func (g *Generator) generateMutationHook(service Service, method Method) (string, error) {
    tmplStr := `
export function use{{.MethodName}}() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async ({{.Params}}) => {
      return window.go.{{.ServiceName}}.{{.MethodName}}({{.ParamNames}});
    },
    {{- if .Optimistic}}
    onMutate: async (variables) => {
      // Optimistic update
      {{- range .InvalidateQueries}}
      await queryClient.cancelQueries({ queryKey: {{.}} });
      {{- end}}
      
      const previousData = queryClient.getQueryData({{.OptimisticKey}});
      
      queryClient.setQueryData({{.OptimisticKey}}, (old) => {
        // Apply optimistic update
        return { ...old, ...variables };
      });
      
      return { previousData };
    },
    onError: (err, variables, context) => {
      // Rollback on error
      if (context?.previousData) {
        queryClient.setQueryData({{.OptimisticKey}}, context.previousData);
      }
    },
    {{- end}}
    onSuccess: (data, variables) => {
      {{- range .InvalidateQueries}}
      queryClient.invalidateQueries({ queryKey: {{.}} });
      {{- end}}
      {{- range .Events}}
      // Emit events for real-time sync
      EventsEmit('{{.}}', data);
      {{- end}}
    },
  });
}
`
    
    // Parse mutation annotation
    mutAnn := extractMutationAnnotation(method.Annotations)
    
    // Generate template data
    data := map[string]interface{}{
        "MethodName":        method.Name,
        "ServiceName":       service.Name,
        "Params":            formatHookParams(method.Params),
        "ParamNames":        formatParamNames(method.Params),
        "Optimistic":        mutAnn.Optimistic,
        "OptimisticKey":     generateOptimisticKey(service.Name, method.Name),
        "InvalidateQueries": formatInvalidateQueries(mutAnn.Invalidate),
        "Events":            mutAnn.Events,
    }
    
    var buf bytes.Buffer
    tmpl := template.Must(template.New("mutation").Parse(tmplStr))
    if err := tmpl.Execute(&buf, data); err != nil {
        return "", err
    }
    
    return buf.String(), nil
}
```

### 4. Type Generation with Validation (`types.go`)
```go
package wrf

import (
    "bytes"
    "fmt"
    "strings"
)

// generateTypeWithValidation generates TypeScript type with Zod schema
func (g *Generator) generateTypeWithValidation(typ Type) (string, string, error) {
    var typeBuf, schemaBuf bytes.Buffer
    
    // Generate TypeScript interface
    typeBuf.WriteString(fmt.Sprintf("export interface %s {\n", typ.Name))
    
    // Generate Zod schema
    schemaBuf.WriteString(fmt.Sprintf("export const %sSchema = z.object({\n", typ.Name))
    
    for _, field := range typ.Fields {
        // TypeScript field
        tsType := g.goTypeToTS(field.Type)
        optional := field.Tag.Get("json") == "omitempty"
        
        if optional {
            typeBuf.WriteString(fmt.Sprintf("  %s?: %s;\n", field.Name, tsType))
        } else {
            typeBuf.WriteString(fmt.Sprintf("  %s: %s;\n", field.Name, tsType))
        }
        
        // Zod validation
        zodType := g.goTypeToZod(field.Type)
        validation := g.extractValidation(field)
        
        if optional {
            schemaBuf.WriteString(fmt.Sprintf("  %s: %s.optional(),\n", field.Name, zodType))
        } else {
            schemaBuf.WriteString(fmt.Sprintf("  %s: %s%s,\n", field.Name, zodType, validation))
        }
    }
    
    typeBuf.WriteString("}\n")
    schemaBuf.WriteString("});\n")
    
    // Add type inference
    schemaBuf.WriteString(fmt.Sprintf("\nexport type %s = z.infer<typeof %sSchema>;\n", 
        typ.Name, typ.Name))
    
    return typeBuf.String(), schemaBuf.String(), nil
}

// extractValidation extracts validation rules from struct tags
func (g *Generator) extractValidation(field Field) string {
    var validations []string
    
    // Parse wrf tags
    if wrfTag := field.Tag.Get("wrf"); wrfTag != "" {
        parts := strings.Split(wrfTag, ",")
        for _, part := range parts {
            switch {
            case part == "required":
                // Already handled by optional check
            case strings.HasPrefix(part, "min:"):
                min := strings.TrimPrefix(part, "min:")
                validations = append(validations, fmt.Sprintf(".min(%s)", min))
            case strings.HasPrefix(part, "max:"):
                max := strings.TrimPrefix(part, "max:")
                validations = append(validations, fmt.Sprintf(".max(%s)", max))
            case strings.HasPrefix(part, "validation:"):
                val := strings.TrimPrefix(part, "validation:")
                switch val {
                case "email":
                    validations = append(validations, ".email()")
                case "url":
                    validations = append(validations, ".url()")
                case "uuid":
                    validations = append(validations, ".uuid()")
                }
            }
        }
    }
    
    return strings.Join(validations, "")
}
```

### 5. Dev Server Extensions (`dev_server.go`)
```go
package wrf

import (
    "encoding/json"
    "net/http"
    "github.com/wailsapp/wails/v2/pkg/plugins"
)

// ConfigureDevServer adds WRF development tools
func (p *WRFPlugin) ConfigureDevServer(config plugins.DevServerConfigurator) error {
    // Add service inspector
    config.AddHandler("/inspector", http.HandlerFunc(p.serveInspector))
    
    // Add type browser
    config.AddHandler("/types", http.HandlerFunc(p.serveTypeBrowser))
    
    // Add event monitor WebSocket
    config.AddWebSocketHandler("/events", p)
    
    // Add React Query devtools bridge
    config.AddHandler("/devtools", http.HandlerFunc(p.serveDevTools))
    
    return nil
}

// serveInspector provides service inspection UI
func (p *WRFPlugin) serveInspector(w http.ResponseWriter, r *http.Request) {
    services := p.parser.GetCachedServices()
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "services": services,
        "stats": map[string]interface{}{
            "totalServices": len(services),
            "totalMethods":  countMethods(services),
            "lastGenerated": p.generator.LastGenerated(),
        },
    })
}

// HandleWebSocket for event monitoring
func (p *WRFPlugin) HandleWebSocket(ctx *plugins.WebSocketContext) error {
    p.logger.Info("WRF event monitor connected")
    
    // Subscribe to all events for monitoring
    unsubscribe := plugins.EventsOnMultiple(ctx.Conn, "*", func(data ...interface{}) {
        event := map[string]interface{}{
            "type":      "event",
            "eventName": data[0],
            "data":      data[1:],
            "timestamp": time.Now().Unix(),
        }
        
        ctx.Conn.WriteJSON(event)
    })
    
    defer unsubscribe()
    
    // Keep connection alive
    for {
        _, _, err := ctx.Conn.ReadMessage()
        if err != nil {
            break
        }
    }
    
    return nil
}
```

## Configuration Example
```json
{
  "plugins": {
    "wrf": {
      "enabled": true,
      "config": {
        "go": {
          "sourcePaths": ["./app/**/*.go"],
          "watchMode": true
        },
        "generation": {
          "hooks": {
            "output": "frontend/src/hooks/generated",
            "framework": "react-query",
            "includeComments": true
          },
          "types": {
            "output": "frontend/src/types/generated.ts",
            "validation": "zod",
            "includeSchemas": true
          },
          "events": {
            "output": "frontend/src/events/generated.ts",
            "typed": true
          }
        },
        "development": {
          "hmr": true,
          "devtools": true,
          "eventDebugging": true
        }
      }
    }
  }
}
```

## Testing Strategy
- Test annotation parsing
- Test React Query hook generation
- Test Zod schema generation  
- Test dev server endpoints
- Integration test with real React app

## Next Steps
With WRF core complete, implement file watching and hot reload integration.