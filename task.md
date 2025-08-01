# Task: Wails Options Integration
**Generated from Master Planning**: 2024-01-15
**Context Package**: `/requests/wails-plugin-fork/context/`

## Task Sizing Assessment
**File Count**: 2-3 files
**Estimated Time**: 30 minutes
**Token Estimate**: 40k tokens
**Complexity Level**: 2 (Moderate)
**Parallelization Benefit**: MEDIUM - Can be done alongside CLI work
**Atomicity Assessment**: ✅ ATOMIC - Configuration extension
**Boundary Analysis**: ✅ CLEAR - Limited to options package

## Persona Assignment
**Persona**: Software Engineer
**Expertise Required**: Go structs, JSON marshaling, configuration patterns
**Worktree**: `~/work/wrf/wails-plugin-fork/`

## Context Summary
**Risk Level**: MEDIUM - Modifies existing Wails types
**Integration Points**: Used by all components reading configuration
**Architecture Pattern**: Extend existing options pattern
**Similar Reference**: Existing Wails options structure

### Task Scope Boundaries
**MODIFY Zone** (Direct Changes):
```yaml
primary_files:
  - /v2/pkg/options/app.go               # Add plugin configuration
  - /v2/pkg/options/plugin.go            # New plugin-specific options
  - /v2/pkg/options/app_test.go          # Update tests
```

**REVIEW Zone** (Check for Impact):
```yaml
check_usage:
  - /v2/cmd/wails/internal/commands/     # Command usage of options
  - /v2/internal/app/                    # App initialization
```

**IGNORE Zone** (Do Not Touch):
```yaml
ignore_completely:
  - /v2/pkg/options/assetserver.go       # Other option types
  - /v2/pkg/options/linux.go             # Platform-specific options
```

## Task Requirements
**Objective**: Extend Wails options to support plugin configuration while maintaining backward compatibility

**Success Criteria**:
- [ ] Plugin configuration added to App options
- [ ] JSON marshaling/unmarshaling works correctly
- [ ] Backward compatibility maintained (plugins optional)
- [ ] Validation for plugin configuration
- [ ] Documentation updated
- [ ] Tests pass with and without plugins

**Validation Commands**:
```bash
go test ./v2/pkg/options/...             # All tests pass
go build ./v2/pkg/options/...            # No compilation errors
# Test backward compatibility
echo '{}' | go run ./tools/checkconfig   # Empty config still valid
```

## Implementation Details

### 1. Extend App Options (`app.go`)
```go
// Add to existing App struct
type App struct {
    // ... existing fields ...
    
    // Plugin system configuration (optional)
    Plugins *PluginSystemConfig `json:"plugins,omitempty"`
}

// Validation method update
func (a *App) Validate() error {
    // ... existing validation ...
    
    if a.Plugins != nil {
        if err := a.Plugins.Validate(); err != nil {
            return fmt.Errorf("plugin configuration error: %w", err)
        }
    }
    
    return nil
}
```

### 2. Plugin Options (`plugin.go`)
```go
package options

import (
    "fmt"
    "time"
)

// PluginSystemConfig configures the plugin system
type PluginSystemConfig struct {
    // Enable/disable entire plugin system
    Enabled bool `json:"enabled"`
    
    // Plugin discovery paths (optional)
    DiscoveryPaths []string `json:"discoveryPaths,omitempty"`
    
    // Load timeout for plugins
    LoadTimeout time.Duration `json:"loadTimeout,omitempty"`
    
    // Individual plugin configurations
    Plugins map[string]*PluginConfig `json:"plugins,omitempty"`
}

// PluginConfig configures an individual plugin
type PluginConfig struct {
    // Enable/disable this specific plugin
    Enabled bool `json:"enabled"`
    
    // Plugin version constraint (optional)
    Version string `json:"version,omitempty"`
    
    // Plugin-specific configuration
    Config map[string]interface{} `json:"config,omitempty"`
    
    // Path for external plugins (optional)
    Path string `json:"path,omitempty"`
}

// Validate plugin configuration
func (p *PluginSystemConfig) Validate() error {
    if p.LoadTimeout < 0 {
        return fmt.Errorf("loadTimeout cannot be negative")
    }
    
    // Set defaults
    if p.LoadTimeout == 0 {
        p.LoadTimeout = 30 * time.Second
    }
    
    // Validate individual plugins
    for name, config := range p.Plugins {
        if name == "" {
            return fmt.Errorf("plugin name cannot be empty")
        }
        if config == nil {
            return fmt.Errorf("plugin %s has nil configuration", name)
        }
    }
    
    return nil
}

// SetDefaults applies default values
func (p *PluginSystemConfig) SetDefaults() {
    if p.LoadTimeout == 0 {
        p.LoadTimeout = 30 * time.Second
    }
    
    if p.DiscoveryPaths == nil {
        p.DiscoveryPaths = []string{
            "./.wails/plugins",
            "~/.wails/plugins",
        }
    }
}
```

### 3. Configuration Example
```json
{
  "name": "MyApp",
  "plugins": {
    "enabled": true,
    "discoveryPaths": [
      "./plugins",
      "~/.wails/plugins"
    ],
    "loadTimeout": "30s",
    "plugins": {
      "wrf": {
        "enabled": true,
        "version": "^1.0.0",
        "config": {
          "go": {
            "sourcePaths": ["./app/**/*.go"]
          },
          "generation": {
            "hooks": {
              "output": "frontend/src/hooks/generated"
            }
          }
        }
      },
      "basic-vue": {
        "enabled": false
      }
    }
  }
}
```

## Testing Requirements
- Test JSON marshaling/unmarshaling
- Test validation with various configurations
- Test backward compatibility (nil plugins)
- Test default values
- Benchmark configuration parsing

## Migration Considerations
- Existing apps without plugins field continue to work
- Clear documentation on plugin configuration
- Example configurations provided

## Next Steps
With configuration in place, we can integrate the plugin manager into the application initialization.