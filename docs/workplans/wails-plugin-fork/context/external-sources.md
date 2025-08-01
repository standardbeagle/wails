# External Context Sources - Wails Plugin Fork

## Primary Documentation

### Wails Documentation
- **Official Docs**: [wails.io](https://wails.io) - Architecture overview, contribution guidelines
- **API Reference**: [pkg.go.dev/github.com/wailsapp/wails/v2](https://pkg.go.dev/github.com/wailsapp/wails/v2) - Package documentation
- **Contributing Guide**: [github.com/wailsapp/wails/blob/master/CONTRIBUTING.md](https://github.com/wailsapp/wails/blob/master/CONTRIBUTING.md) - Code standards, PR process

### Go Plugin Development
- **Go Plugins**: [golang.org/pkg/plugin](https://golang.org/pkg/plugin) - Dynamic loading of Go plugins
- **Interface Design**: [github.com/golang/go/wiki/CodeReviewComments](https://github.com/golang/go/wiki/CodeReviewComments) - Go interface best practices
- **Build Modes**: [golang.org/cmd/go/#hdr-Build_modes](https://golang.org/cmd/go/#hdr-Build_modes) - Plugin build mode documentation

### TypeScript Code Generation
- **TypeScript Compiler API**: [github.com/microsoft/TypeScript/wiki/Using-the-Compiler-API](https://github.com/microsoft/TypeScript/wiki/Using-the-Compiler-API) - AST manipulation
- **Go AST**: [golang.org/pkg/go/ast](https://golang.org/pkg/go/ast) - Parsing Go code for type extraction
- **Code Generation**: [github.com/dave/jennifer](https://github.com/dave/jennifer) - Go code generation library

## Framework Integration Patterns

### React Integration
- **React DevTools Protocol**: [github.com/facebook/react/tree/main/packages/react-devtools](https://github.com/facebook/react/tree/main/packages/react-devtools) - Integration patterns
- **React Query**: [tanstack.com/query](https://tanstack.com/query) - Caching and state management patterns
- **React Refresh**: [github.com/facebook/react/tree/main/packages/react-refresh](https://github.com/facebook/react/tree/main/packages/react-refresh) - HMR implementation

### Vue Integration (Simple Example)
- **Vue Devtools API**: [github.com/vuejs/devtools](https://github.com/vuejs/devtools) - Plugin integration
- **Vue Composition API**: [vuejs.org/guide/extras/composition-api-faq](https://vuejs.org/guide/extras/composition-api-faq) - Hook patterns

## Standards and Best Practices

### Security Standards
- **OWASP Go Security**: [cheatsheetseries.owasp.org/cheatsheets/Go_Security_Cheat_Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Go_Security_Cheat_Sheet) - Secure plugin loading
- **Plugin Isolation**: Sandboxing considerations for external plugins

### Testing Standards
- **Go Testing**: [golang.org/pkg/testing](https://golang.org/pkg/testing) - Standard testing patterns
- **Testify**: [github.com/stretchr/testify](https://github.com/stretchr/testify) - Assertion library used by Wails
- **Mock Generation**: [github.com/golang/mock](https://github.com/golang/mock) - Interface mocking

### API Design
- **REST Guidelines**: While not REST, similar principles for plugin APIs
- **Semantic Versioning**: [semver.org](https://semver.org) - Plugin version compatibility

## Reference Implementations

### Plugin Systems
- **Terraform Providers**: [terraform.io/plugin](https://terraform.io/plugin) - Go plugin architecture reference
- **Cobra Extensions**: [github.com/spf13/cobra](https://github.com/spf13/cobra) - CLI extension patterns
- **Caddy Modules**: [caddyserver.com/docs/extending-caddy](https://caddyserver.com/docs/extending-caddy) - Web server plugin system

### Code Generation Examples
- **Wire**: [github.com/google/wire](https://github.com/google/wire) - Dependency injection code generation
- **Swag**: [github.com/swaggo/swag](https://github.com/swaggo/swag) - Annotation-based code generation
- **Jennifer**: [github.com/dave/jennifer](https://github.com/dave/jennifer) - Go code generation

## Development Tools

### Build Tools
- **Mage**: [magefile.org](https://magefile.org) - Build tool that could manage plugin builds
- **GoReleaser**: [goreleaser.com](https://goreleaser.com) - For plugin distribution

### Debugging
- **Delve**: [github.com/go-delve/delve](https://github.com/go-delve/delve) - Go debugger for plugin development
- **pprof**: [golang.org/pkg/net/http/pprof](https://golang.org/pkg/net/http/pprof) - Performance profiling