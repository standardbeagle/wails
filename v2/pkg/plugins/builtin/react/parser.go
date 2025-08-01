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

// Param represents a method parameter
type Param struct {
	Name string
	Type string
}

// Result represents a method return value
type Result struct {
	Name string
	Type string
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

// extractDoc extracts documentation from comment group
func extractDoc(comments *ast.CommentGroup) string {
	if comments == nil {
		return ""
	}
	
	var lines []string
	for _, comment := range comments.List {
		text := strings.TrimPrefix(comment.Text, "//")
		text = strings.TrimPrefix(text, "/*")
		text = strings.TrimSuffix(text, "*/")
		text = strings.TrimSpace(text)
		if text != "" {
			lines = append(lines, text)
		}
	}
	
	return strings.Join(lines, " ")
}

// isMethodOf checks if a function is a method of the given type
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

// isExported checks if a name is exported (starts with uppercase)
func isExported(name string) bool {
	return len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z'
}

// typeToString converts an AST type to a string
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