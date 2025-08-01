# Codebase Context Documentation - Wails Plugin Fork

## Wails v2 Architecture Analysis

### Core Package Structure
- **cmd/wails/**: CLI tool implementation
  - `internal/commands/build/`: Build command logic
  - `internal/commands/dev/`: Development server command
  - `internal/commands/init/`: Project initialization
  - Main entry point for all CLI operations

- **pkg/**: Public packages
  - `options/`: Application configuration structures
  - `assetserver/`: Static asset serving
  - `templates/`: Project templates
  - `logger/`: Logging utilities

- **internal/**: Private implementation
  - `app/`: Core application runtime
  - `binding/`: Go-JS binding generation
  - `frontend/`: Platform-specific frontend implementations
  - `fs/`: Filesystem utilities
  - `shell/`: Command execution

### Key Integration Points for Plugin System

1. **Build Process Hook Points**
   - File: `cmd/wails/internal/commands/build/build.go`
   - Integration: Add plugin hooks before/after build steps
   - Pattern: Command pattern with context passing

2. **Development Server Integration**
   - File: `cmd/wails/internal/commands/dev/dev.go`
   - Integration: Add middleware and WebSocket handler registration
   - Pattern: HTTP handler chain with plugin extensions

3. **Binding Generation Extension**
   - File: `internal/binding/binding.go`
   - Integration: Allow plugins to participate in code generation
   - Pattern: Visitor pattern for AST traversal

4. **CLI Command Structure**
   - File: `cmd/wails/main.go`
   - Integration: Add plugin subcommands
   - Pattern: Cobra command tree extension

### Minimal Modification Strategy

To maximize upstream merge potential:
1. Add new `pkg/plugins` package without modifying existing packages
2. Use interface injection at key points rather than direct modifications
3. Make plugin system completely opt-in via configuration
4. Follow existing patterns for error handling and logging

### Similar Patterns in Codebase

1. **Template System** (`pkg/templates/`)
   - Similar plugin discovery mechanism
   - Template registration pattern to follow

2. **Asset Server Middleware** (`pkg/assetserver/`)
   - Middleware chain pattern for dev server plugins

3. **Options Pattern** (`pkg/options/`)
   - Configuration structure to extend for plugins

### File Dependency Analysis

High-change areas (new files):
- `/v2/pkg/plugins/`: New plugin system package
- `/v2/cmd/wails/internal/commands/plugin/`: New CLI commands

Medium-change areas (hook additions):
- `/v2/cmd/wails/internal/commands/build/build.go`: Build hooks
- `/v2/cmd/wails/internal/commands/dev/dev.go`: Dev server hooks
- `/v2/pkg/options/app.go`: Plugin configuration

Low-change areas (minimal touches):
- `/v2/cmd/wails/main.go`: Command registration
- `/v2/internal/binding/binding.go`: Extension point