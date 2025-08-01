package wrf

import (
	"fmt"
	"strings"
	"text/template"
	"time"
)

// Template helper functions and formatting methods

// templateFuncs returns template functions for code generation
func (g *Generator) templateFuncs() template.FuncMap {
	return template.FuncMap{
		"lowerFirst":    lowerFirst,
		"formatDuration": formatDuration,
		"join":          strings.Join,
		"quote":         func(s string) string { return fmt.Sprintf(`"%s"`, s) },
	}
}

// formatHookParams formats method parameters for TypeScript hooks
func (g *Generator) formatHookParams(params []Param) string {
	if len(params) == 0 {
		return ""
	}

	var parts []string
	for _, param := range params {
		tsType := g.goTypeToTS(param.Type)
		parts = append(parts, fmt.Sprintf("%s: %s", param.Name, tsType))
	}

	return strings.Join(parts, ", ")
}

// formatParamNames formats parameter names for function calls
func (g *Generator) formatParamNames(params []Param) string {
	if len(params) == 0 {
		return ""
	}

	var names []string
	for _, param := range params {
		names = append(names, param.Name)
	}

	return strings.Join(names, ", ")
}

// formatReturns formats return types for TypeScript
func (g *Generator) formatReturns(results []Result) string {
	if len(results) == 0 {
		return "void"
	}

	if len(results) == 1 {
		return g.goTypeToTS(results[0].Type)
	}

	// Multiple return values - create a tuple
	var parts []string
	for _, result := range results {
		parts = append(parts, g.goTypeToTS(result.Type))
	}

	return fmt.Sprintf("[%s]", strings.Join(parts, ", "))
}

// generateQueryKey generates a query key for React Query
func (g *Generator) generateQueryKey(serviceName, methodName string, params []Param) string {
	if len(params) == 0 {
		return fmt.Sprintf(`["%s", "%s"]`, serviceName, lowerFirst(methodName))
	}

	paramNames := make([]string, len(params))
	for i, param := range params {
		paramNames[i] = param.Name
	}

	return fmt.Sprintf(`["%s", "%s", %s]`, serviceName, lowerFirst(methodName), strings.Join(paramNames, ", "))
}

// generateOptimisticKey generates an optimistic update key
func (g *Generator) generateOptimisticKey(serviceName, methodName string) string {
	return fmt.Sprintf(`["%s", "%s"]`, serviceName, lowerFirst(methodName))
}

// formatInvalidateQueries formats query invalidation patterns
func (g *Generator) formatInvalidateQueries(patterns []string) []string {
	var formatted []string
	for _, pattern := range patterns {
		// Convert pattern to query key format
		parts := strings.Split(pattern, ":")
		if len(parts) == 2 {
			formatted = append(formatted, fmt.Sprintf(`["%s", "%s"]`, parts[0], parts[1]))
		} else {
			formatted = append(formatted, fmt.Sprintf(`["%s"]`, pattern))
		}
	}
	return formatted
}

// goTypeToTS converts Go types to TypeScript types
func (g *Generator) goTypeToTS(goType string) string {
	// Handle common Go types
	switch goType {
	case "string":
		return "string"
	case "int", "int8", "int16", "int32", "int64", 
		 "uint", "uint8", "uint16", "uint32", "uint64",
		 "float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	case "[]byte":
		return "Uint8Array"
	case "error":
		return "Error"
	case "interface{}":
		return "any"
	}

	// Handle arrays
	if strings.HasPrefix(goType, "[]") {
		elemType := strings.TrimPrefix(goType, "[]")
		return fmt.Sprintf("%s[]", g.goTypeToTS(elemType))
	}

	// Handle pointers
	if strings.HasPrefix(goType, "*") {
		elemType := strings.TrimPrefix(goType, "*")
		if g.config.Generation.Types.NullableFields {
			return fmt.Sprintf("%s | null", g.goTypeToTS(elemType))
		}
		return g.goTypeToTS(elemType)
	}

	// Handle maps
	if strings.HasPrefix(goType, "map[") {
		return "Record<string, any>"
	}

	// Handle time.Time
	if strings.Contains(goType, "time.Time") {
		return "Date"
	}

	// Handle context.Context
	if strings.Contains(goType, "context.Context") {
		return "any" // Context is handled by the runtime
	}

	// Default to any for unknown types, but preserve the original name
	// This allows for custom types to be generated properly
	return goType
}

// goTypeToZod converts Go types to Zod validation types
func (g *Generator) goTypeToZod(goType string) string {
	switch goType {
	case "string":
		return "z.string()"
	case "int", "int8", "int16", "int32", "int64",
		 "uint", "uint8", "uint16", "uint32", "uint64":
		return "z.number().int()"
	case "float32", "float64":
		return "z.number()"
	case "bool":
		return "z.boolean()"
	case "[]byte":
		return "z.instanceof(Uint8Array)"
	case "interface{}":
		return "z.any()"
	}

	// Handle arrays
	if strings.HasPrefix(goType, "[]") {
		elemType := strings.TrimPrefix(goType, "[]")
		return fmt.Sprintf("z.array(%s)", g.goTypeToZod(elemType))
	}

	// Handle pointers
	if strings.HasPrefix(goType, "*") {
		elemType := strings.TrimPrefix(goType, "*")
		return fmt.Sprintf("%s.nullable()", g.goTypeToZod(elemType))
	}

	// Handle maps
	if strings.HasPrefix(goType, "map[") {
		return "z.record(z.any())"
	}

	// Handle time.Time
	if strings.Contains(goType, "time.Time") {
		return "z.date()"
	}

	// Default to any for complex types
	return "z.any()"
}

// isFieldOptional determines if a field should be optional in TypeScript
func (g *Generator) isFieldOptional(field Field) bool {
	if field.Tag != nil {
		jsonTag := field.Tag.Get("json")
		if strings.Contains(jsonTag, "omitempty") {
			return true
		}
	}

	// Check for pointer types
	if strings.HasPrefix(field.Type, "*") {
		return true
	}

	return false
}

// fieldNameToTS converts Go field names to TypeScript property names
func (g *Generator) fieldNameToTS(fieldName string) string {
	// Convert Go field names to camelCase for TypeScript
	if len(fieldName) == 0 {
		return fieldName
	}

	// If already camelCase, return as-is
	if fieldName[0] >= 'a' && fieldName[0] <= 'z' {
		return fieldName
	}

	// Convert first character to lowercase
	return strings.ToLower(fieldName[:1]) + fieldName[1:]
}

// extractValidation extracts validation rules from struct tags
func (g *Generator) extractValidation(field Field) string {
	if field.Tag == nil {
		return ""
	}

	var validations []string

	// Parse wrf tags for validation
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

// Utility functions

func lowerFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func formatDuration(d time.Duration) string {
	return fmt.Sprintf("%d", int(d.Milliseconds()))
}