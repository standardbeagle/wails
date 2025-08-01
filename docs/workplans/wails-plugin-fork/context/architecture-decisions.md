# Architecture Decisions - Wails Plugin Fork

## Core Design Decisions

### 1. Plugin Architecture: Interface-Based vs Binary Plugins
**Decision**: Interface-based plugins with optional binary (.so) support
**Rationale**: 
- Interface-based allows easier testing and upstream merge
- Binary plugins optional for third-party extensions
- Maintains Go's type safety and compile-time checks
**Trade-offs**: 
- (+) Better performance, easier debugging
- (+) Simpler for upstream merge
- (-) Requires recompilation for new plugins
**Alternative**: Pure binary plugins (rejected due to complexity)

### 2. Integration Strategy: Minimal Core Changes
**Decision**: Add hooks via interface injection, new packages only
**Rationale**:
- Maximizes chance of upstream acceptance
- Maintains backward compatibility
- Easier to maintain fork sync
**Implementation**:
- New `pkg/plugins` package
- Interface injection at 4-5 key points
- All changes behind feature flags

### 3. Configuration Approach: Extend Existing System
**Decision**: Extend wails.json with plugins section
**Rationale**:
- Follows existing patterns
- Single configuration file
- Familiar to Wails users
**Format**:
```json
{
  "plugins": {
    "enabled": true,
    "plugins": {
      "wrf": {
        "enabled": true,
        "config": {}
      }
    }
  }
}
```

### 4. Code Generation Strategy: AST-Based
**Decision**: Parse Go AST, generate TypeScript via templates
**Rationale**:
- Accurate type extraction
- Flexible output generation
- Supports annotations for metadata
**Tools**:
- Go's ast package for parsing
- Template-based generation
- Optional: Jennifer for Go code generation

### 5. Build Integration: Hook-Based
**Decision**: Pre/post build hooks at key stages
**Rationale**:
- Non-invasive to existing build process
- Clear execution order
- Easy to disable
**Hook Points**:
1. Pre-build validation
2. Post-Go analysis
3. Pre-frontend build
4. Post-build optimization

### 6. Development Server: Middleware Chain
**Decision**: Extensible middleware and WebSocket handlers
**Rationale**:
- Follows existing HTTP patterns
- Allows plugin tools/debugging
- Supports real-time features
**Implementation**:
- Middleware registration API
- WebSocket multiplexing
- Plugin-specific routes

### 7. Error Handling: Fail-Safe
**Decision**: Plugin errors should not crash Wails
**Rationale**:
- Stability for upstream merge
- Better developer experience
- Graceful degradation
**Strategy**:
- Recover from panics
- Log errors clearly
- Continue without failed plugins

### 8. Testing Strategy: Comprehensive
**Decision**: Unit tests, integration tests, example plugins
**Rationale**:
- Required for upstream merge
- Ensures stability
- Documents usage
**Approach**:
- Mock plugins for testing
- Integration test suite
- Real plugin examples

## Implementation Priorities

1. **Phase 1**: Core plugin system (interfaces, manager, lifecycle)
2. **Phase 2**: Build integration (hooks, code generation)
3. **Phase 3**: Dev server enhancement (middleware, WebSocket)
4. **Phase 4**: React plugin implementation (WRF)

## Risk Mitigation Strategies

### Upstream Merge Risk
- Keep changes minimal and clean
- Follow Wails coding standards exactly
- Maintain comprehensive tests
- Document thoroughly

### Performance Risk
- Lazy plugin loading
- Caching for code generation
- Benchmarks to prove no regression

### Compatibility Risk
- Feature flags for all plugin features
- No changes to existing APIs
- Extensive compatibility testing