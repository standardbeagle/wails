package wrf

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/plugins"
)

// Parser handles Go AST parsing for WRF
type Parser struct {
	config *Config
	logger plugins.Logger
}

// NewParser creates a new parser instance
func NewParser(config *Config, logger plugins.Logger) *Parser {
	return &Parser{
		config: config,
		logger: logger,
	}
}

// Service represents a parsed Go service with annotations
type Service struct {
	Name        string
	Methods     []Method
	Doc         string
	Annotations []Annotation
}

// Method represents a parsed Go method with annotations
type Method struct {
	Name        string
	Params      []Param
	Results     []Result
	Doc         string
	Receiver    string
	Annotations []Annotation
}

// Param represents a method parameter
type Param struct {
	Name string
	Type string
	Tag  string
}

// Result represents a method return value
type Result struct {
	Name string
	Type string
}

// Type represents a parsed Go type
type Type struct {
	Name        string
	Fields      []Field
	Doc         string
	Annotations []Annotation
}

// Field represents a struct field
type Field struct {
	Name        string
	Type        string
	Tag         *StructTag
	Doc         string
	Annotations []Annotation
}

// StructTag represents a struct tag
type StructTag struct {
	Raw string
}

// Get returns the value associated with key in the tag string
func (tag *StructTag) Get(key string) string {
	if tag == nil {
		return ""
	}
	
	// Simple tag parsing - extract value for key
	parts := strings.Split(tag.Raw, " ")
	for _, part := range parts {
		if strings.HasPrefix(part, key+":") {
			value := strings.TrimPrefix(part, key+":")
			value = strings.Trim(value, `"`)
			return value
		}
	}
	
	return ""
}

// ParseServices parses Go services with WRF annotations
func (p *Parser) ParseServices(projectRoot string) ([]Service, error) {
	var services []Service

	for _, sourcePath := range p.config.Go.SourcePaths {
		pattern := filepath.Join(projectRoot, sourcePath)
		files, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			fileServices, err := p.parseServicesFromFile(file)
			if err != nil {
				p.logger.Warn("Failed to parse services from file", "file", file, "error", err)
				continue
			}
			services = append(services, fileServices...)
		}
	}

	return services, nil
}

// ParseTypes parses Go types for generation
func (p *Parser) ParseTypes(projectRoot string) ([]Type, error) {
	var types []Type

	for _, sourcePath := range p.config.Go.SourcePaths {
		pattern := filepath.Join(projectRoot, sourcePath)
		files, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			fileTypes, err := p.parseTypesFromFile(file)
			if err != nil {
				p.logger.Warn("Failed to parse types from file", "file", file, "error", err)
				continue
			}
			types = append(types, fileTypes...)
		}
	}

	return types, nil
}

// parseServicesFromFile parses services from a single Go file
func (p *Parser) parseServicesFromFile(filename string) ([]Service, error) {
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

			// Extract service annotations
			comments := extractComments(genDecl.Doc)
			annotations := ParseAnnotations(comments)

			service := Service{
				Name:        typeSpec.Name.Name,
				Doc:         strings.Join(comments, " "),
				Annotations: annotations,
			}

			// Find methods
			for _, decl := range node.Decls {
				if funcDecl, ok := decl.(*ast.FuncDecl); ok {
					if isMethodOf(funcDecl, typeSpec.Name.Name) {
						method := p.parseMethodWithAnnotations(funcDecl)
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

// parseTypesFromFile parses types from a single Go file
func (p *Parser) parseTypesFromFile(filename string) ([]Type, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var types []Type

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

			// Only process struct types
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			// Extract type annotations
			comments := extractComments(genDecl.Doc)
			annotations := ParseAnnotations(comments)

			typ := Type{
				Name:        typeSpec.Name.Name,
				Doc:         strings.Join(comments, " "),
				Annotations: annotations,
			}

			// Parse struct fields
			for _, field := range structType.Fields.List {
				for _, name := range field.Names {
					if isExported(name.Name) {
						fieldComments := extractComments(field.Doc)
						fieldAnnotations := ParseAnnotations(fieldComments)

						var tag *StructTag
						if field.Tag != nil {
							tag = &StructTag{Raw: field.Tag.Value}
						}

						typ.Fields = append(typ.Fields, Field{
							Name:        name.Name,
							Type:        typeToString(field.Type),
							Tag:         tag,
							Doc:         strings.Join(fieldComments, " "),
							Annotations: fieldAnnotations,
						})
					}
				}
			}

			if len(typ.Fields) > 0 {
				types = append(types, typ)
			}
		}
	}

	return types, nil
}

// parseMethodWithAnnotations parses a method with WRF annotations
func (p *Parser) parseMethodWithAnnotations(funcDecl *ast.FuncDecl) *Method {
	comments := extractComments(funcDecl.Doc)
	annotations := ParseAnnotations(comments)

	method := &Method{
		Name:        funcDecl.Name.Name,
		Doc:         strings.Join(comments, " "),
		Annotations: annotations,
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

// Helper functions

func extractComments(commentGroup *ast.CommentGroup) []string {
	if commentGroup == nil {
		return nil
	}

	var comments []string
	for _, comment := range commentGroup.List {
		text := strings.TrimPrefix(comment.Text, "//")
		text = strings.TrimPrefix(text, "/*")
		text = strings.TrimSuffix(text, "*/")
		text = strings.TrimSpace(text)
		if text != "" {
			comments = append(comments, text)
		}
	}

	return comments
}

func isMethodOf(funcDecl *ast.FuncDecl, typeName string) bool {
	if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
		return false
	}

	recvType := funcDecl.Recv.List[0].Type

	// Handle pointer receivers
	if starExpr, ok := recvType.(*ast.StarExpr); ok {
		recvType = starExpr.X
	}

	if ident, ok := recvType.(*ast.Ident); ok {
		return ident.Name == typeName
	}

	return false
}

func isExported(name string) bool {
	return len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z'
}

func typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + typeToString(t.X)
	case *ast.ArrayType:
		return "[]" + typeToString(t.Elt)
	case *ast.SelectorExpr:
		return typeToString(t.X) + "." + t.Sel.Name
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.MapType:
		return "map[" + typeToString(t.Key) + "]" + typeToString(t.Value)
	case *ast.ChanType:
		return "chan " + typeToString(t.Value)
	case *ast.FuncType:
		return "func"
	default:
		return "unknown"
	}
}