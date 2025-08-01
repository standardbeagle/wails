# Task: Basic React Plugin Implementation
**Generated from Master Planning**: 2024-01-15
**Context Package**: `/requests/wails-plugin-fork/context/`

## Task Sizing Assessment
**File Count**: 5-6 files
**Estimated Time**: 60 minutes
**Token Estimate**: 120k tokens
**Complexity Level**: 3 (Complex)
**Parallelization Benefit**: HIGH - Independent of other plugins
**Atomicity Assessment**: ✅ ATOMIC - Complete plugin example
**Boundary Analysis**: ✅ CLEAR - Self-contained plugin

## Persona Assignment
**Persona**: Software Engineer
**Expertise Required**: Go, TypeScript, React, code generation
**Worktree**: `~/work/wrf/wails-plugin-fork/`

## Context Summary
**Risk Level**: MEDIUM - First real plugin implementation
**Integration Points**: Uses all plugin interfaces
**Architecture Pattern**: Demonstrates plugin capabilities
**Similar Reference**: WRF plugin design document

### Task Scope Boundaries
**MODIFY Zone** (Direct Changes):
```yaml
primary_files:
  - /v2/pkg/plugins/builtin/react/plugin.go          # Main plugin implementation
  - /v2/pkg/plugins/builtin/react/generator.go       # Code generation logic
  - /v2/pkg/plugins/builtin/react/templates.go       # TypeScript templates
  - /v2/pkg/plugins/builtin/react/config.go          # Configuration handling
  - /v2/pkg/plugins/builtin/react/parser.go          # Go AST parsing
  - /v2/pkg/plugins/builtin/react/react_test.go      # Tests
```

**REVIEW Zone** (Check for Impact):
```yaml
check_patterns:
  - /v2/pkg/plugins/builtin.go           # Register plugin
  - /v2/internal/binding/                # Binding generation patterns
```

**IGNORE Zone** (Do Not Touch):
```yaml
ignore_completely:
  - /v2/pkg/plugins/builtin/vue/         # Other framework plugins
  - /v2/cmd/                             # CLI already done
```

## Task Requirements
**Objective**: Implement a basic React plugin demonstrating TypeScript generation and hooks

**Success Criteria**:
- [ ] Generates TypeScript types from Go structs
- [ ] Creates React hooks for Go methods
- [ ] Supports basic type mapping
- [ ] Handles errors gracefully
- [ ] Includes configuration options
- [ ] Well-documented and tested

**Validation Commands**:
```bash
# Test plugin listing
wails plugin list                        # Should show basic-react

# Test code generation
wails plugin generate                    # Should create TS files

# Verify generated code
ls frontend/src/hooks/generated/         # Should contain hooks
cat frontend/src/types/generated.ts      # Should contain types
```

## Implementation Details

### 1. Main Plugin Implementation (`plugin.go`)
```go
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
    config     *Config
    logger     plugins.Logger
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
        TypesOutput:   "frontend/src/types/generated.ts",
        HooksOutput:   "frontend/src/hooks/generated",
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
```

### 2. Go AST Parser (`parser.go`)
```go
package react

import (
    "go/ast"
    "go/parser"
    "go/token"
    "path/filepath"
    "strings"
)

// Service represents a parsed Go service
type Service struct {
    Name    string
    Methods []Method
    Doc     string
}

// Method represents a parsed Go method
type Method struct {
    Name       string
    Params     []Param
    Results    []Result
    Doc        string
    Receiver   string
}

// parseGoServices parses Go source files for services
func (p *BasicReactPlugin) parseGoServices(projectRoot string) ([]Service, error) {
    var services []Service
    
    // Find Go files in app directory
    pattern := filepath.Join(projectRoot, "app", "*.go")
    files, err := filepath.Glob(pattern)
    if err != nil {
        return nil, err
    }
    
    for _, file := range files {
        fileServices, err := p.parseFile(file)
        if err != nil {
            p.logger.Warn("Failed to parse file", "file", file, "error", err)
            continue
        }
        services = append(services, fileServices...)
    }
    
    return services, nil
}

// parseFile parses a single Go file
func (p *BasicReactPlugin) parseFile(filename string) ([]Service, error) {
    fset := token.NewFileSet()
    node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
    if err != nil {
        return nil, err
    }
    
    var services []Service
    
    // Look for type declarations
    for _, decl := range node.Decls {
        genDecl, ok := decl.(*ast.GenDecl)
        if !ok || genDecl.Tok != token.TYPE {
            continue
        }
        
        for _, spec := range genDecl.Specs {
            typeSpec, ok := spec.(*ast.TypeSpec)
            if !ok {
                continue
            }
            
            // Check if it's a struct
            if _, ok := typeSpec.Type.(*ast.StructType); !ok {
                continue
            }
            
            // Look for methods on this type
            service := Service{
                Name: typeSpec.Name.Name,
                Doc:  extractDoc(genDecl.Doc),
            }
            
            // Find methods
            for _, decl := range node.Decls {
                if funcDecl, ok := decl.(*ast.FuncDecl); ok {
                    if isMethodOf(funcDecl, typeSpec.Name.Name) {
                        method := p.parseMethod(funcDecl)
                        if method != nil && isExported(method.Name) {
                            service.Methods = append(service.Methods, *method)
                        }
                    }
                }
            }
            
            if len(service.Methods) > 0 {
                services = append(services, service)
            }
        }
    }
    
    return services, nil
}

// parseMethod parses a method declaration
func (p *BasicReactPlugin) parseMethod(funcDecl *ast.FuncDecl) *Method {
    method := &Method{
        Name: funcDecl.Name.Name,
        Doc:  extractDoc(funcDecl.Doc),
    }
    
    // Parse parameters
    if funcDecl.Type.Params != nil {
        for _, field := range funcDecl.Type.Params.List {
            param := Param{
                Type: typeToString(field.Type),
            }
            
            // Handle named parameters
            for _, name := range field.Names {
                param.Name = name.Name
                method.Params = append(method.Params, param)
            }
        }
    }
    
    // Parse return values
    if funcDecl.Type.Results != nil {
        for _, field := range funcDecl.Type.Results.List {
            result := Result{
                Type: typeToString(field.Type),
            }
            method.Results = append(method.Results, result)
        }
    }
    
    return method
}
```

### 3. TypeScript Generation (`generator.go`)
```go
package react

import (
    "bytes"
    "fmt"
    "strings"
    "text/template"
)

// generateTypes generates TypeScript type definitions
func (p *BasicReactPlugin) generateTypes(services []Service) (*plugins.GeneratedFile, error) {
    var buf bytes.Buffer
    
    // Write header
    buf.WriteString("// Generated by Wails basic-react plugin\n")
    buf.WriteString("// DO NOT EDIT - This file is automatically generated\n\n")
    
    // Generate interfaces for each service
    for _, service := range services {
        if p.config.IncludeComments && service.Doc != "" {
            buf.WriteString(fmt.Sprintf("/**\n * %s\n */\n", service.Doc))
        }
        
        buf.WriteString(fmt.Sprintf("export interface %sService {\n", service.Name))
        
        for _, method := range service.Methods {
            if p.config.IncludeComments && method.Doc != "" {
                buf.WriteString(fmt.Sprintf("  /**\n   * %s\n   */\n", method.Doc))
            }
            
            // Generate method signature
            params := p.formatParams(method.Params)
            returns := p.formatReturns(method.Results)
            
            buf.WriteString(fmt.Sprintf("  %s(%s): Promise<%s>;\n", 
                lowerFirst(method.Name), params, returns))
        }
        
        buf.WriteString("}\n\n")
    }
    
    // Generate window.go declaration
    buf.WriteString("declare global {\n")
    buf.WriteString("  interface Window {\n")
    buf.WriteString("    go: {\n")
    
    for _, service := range services {
        buf.WriteString(fmt.Sprintf("      %s: %sService;\n", service.Name, service.Name))
    }
    
    buf.WriteString("    };\n")
    buf.WriteString("  }\n")
    buf.WriteString("}\n\n")
    buf.WriteString("export {};\n")
    
    return &plugins.GeneratedFile{
        Path:      p.config.TypesOutput,
        Content:   buf.Bytes(),
        Mode:      0644,
        Overwrite: true,
    }, nil
}

// generateHooks generates React hooks for services
func (p *BasicReactPlugin) generateHooks(services []Service) ([]*plugins.GeneratedFile, error) {
    var files []*plugins.GeneratedFile
    
    hookTemplate := template.Must(template.New("hook").Parse(hookTemplateStr))
    
    for _, service := range services {
        var buf bytes.Buffer
        
        // Generate header
        buf.WriteString("// Generated by Wails basic-react plugin\n")
        buf.WriteString("// DO NOT EDIT - This file is automatically generated\n\n")
        buf.WriteString("import { useState, useCallback } from 'react';\n\n")
        
        // Generate hooks for each method
        for _, method := range service.Methods {
            data := map[string]interface{}{
                "ServiceName": service.Name,
                "MethodName":  method.Name,
                "HookName":    fmt.Sprintf("use%s", method.Name),
                "Params":      p.formatHookParams(method.Params),
                "ParamNames":  p.formatParamNames(method.Params),
                "ReturnType":  p.formatReturns(method.Results),
            }
            
            if err := hookTemplate.Execute(&buf, data); err != nil {
                return nil, err
            }
            
            buf.WriteString("\n")
        }
        
        filename := fmt.Sprintf("%s/%sHooks.ts", p.config.HooksOutput, service.Name)
        files = append(files, &plugins.GeneratedFile{
            Path:      filename,
            Content:   buf.Bytes(),
            Mode:      0644,
            Overwrite: true,
        })
    }
    
    return files, nil
}

// Hook template
const hookTemplateStr = `export function {{.HookName}}() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  
  const execute = useCallback(async ({{.Params}}) => {
    setLoading(true);
    setError(null);
    
    try {
      const result = await window.go.{{.ServiceName}}.{{.MethodName}}({{.ParamNames}});
      return result;
    } catch (err) {
      const error = err instanceof Error ? err : new Error(String(err));
      setError(error);
      throw error;
    } finally {
      setLoading(false);
    }
  }, []);
  
  return { execute, loading, error };
}`
```

### 4. Configuration (`config.go`)
```go
package react

import (
    "fmt"
)

// Config holds plugin configuration
type Config struct {
    TypesOutput     string
    HooksOutput     string
    IncludeComments bool
    SourcePaths     []string
}

// ParseFrom parses configuration from map
func (c *Config) ParseFrom(config map[string]interface{}) error {
    if v, ok := config["typesOutput"].(string); ok {
        c.TypesOutput = v
    }
    
    if v, ok := config["hooksOutput"].(string); ok {
        c.HooksOutput = v
    }
    
    if v, ok := config["includeComments"].(bool); ok {
        c.IncludeComments = v
    }
    
    if v, ok := config["sourcePaths"].([]interface{}); ok {
        for _, path := range v {
            if s, ok := path.(string); ok {
                c.SourcePaths = append(c.SourcePaths, s)
            }
        }
    }
    
    return c.Validate()
}

// Validate configuration
func (c *Config) Validate() error {
    if c.TypesOutput == "" {
        return fmt.Errorf("typesOutput is required")
    }
    
    if c.HooksOutput == "" {
        return fmt.Errorf("hooksOutput is required")
    }
    
    return nil
}
```

## Type Mapping

```go
// Go to TypeScript type mapping
var typeMap = map[string]string{
    "string":     "string",
    "int":        "number",
    "int32":      "number",
    "int64":      "number",
    "float32":    "number",
    "float64":    "number",
    "bool":       "boolean",
    "[]byte":     "Uint8Array",
    "time.Time":  "Date",
    "error":      "Error",
}
```

## Testing Requirements
- Test Go parsing with various service patterns
- Test TypeScript generation correctness
- Test hook generation
- Test configuration validation
- Integration test with real project

## Next Steps
With the basic React plugin complete, we can implement the WRF plugin with advanced features.