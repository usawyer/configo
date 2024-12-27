package yaml

import (
	"fmt"
	"reflect"
	"strings"
)

// fieldInfo represents a single line in the generated YAML template
// along with an optional help (comment) text.
type fieldInfo struct {
	Line string
	Help string
}

// GenerateYAMLTemplate generates a YAML template from a given configuration struct.
// It scans the struct using reflection, collects information about each field,
// and then produces YAML lines aligned with optional help text (comments).
func GenerateYAMLTemplate(cfg interface{}, printDescription bool) string {
	var lines []fieldInfo

	// First pass: Parse the struct and collect the lines
	parseStructure(reflect.TypeOf(cfg), reflect.ValueOf(cfg), 0, &lines)

	// Second pass: Align the resulting YAML lines with help comments
	return generateYAMLWithAlignment(lines, printDescription)
}

// parseStructure recursively traverses a struct (and nested structs)
// to build a list of fieldInfo lines that represent the YAML structure.
func parseStructure(t reflect.Type, v reflect.Value, indent int, lines *[]fieldInfo) {
	indentation := strings.Repeat("  ", indent)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields.
		// In Go, an exported field has an uppercase first letter and an empty PkgPath.
		if field.PkgPath != "" {
			continue
		}

		// Check if the field is intentionally ignored by `yaml:"-"` or `mapstructure:"-"`.
		tag := field.Tag
		if tag.Get("yaml") == "-" || tag.Get("mapstructure") == "-" {
			continue
		}

		// Determine the YAML (and Viper) key name.
		fieldName := getFieldName(field)

		// Retrieve default value (if any).
		defaultValue := getDefaultValue(tag)

		// Retrieve help text (if any).
		helpText := getHelpText(tag)

		switch field.Type.Kind() {
		case reflect.Struct:
			// For nested structs, we append the struct name and recurse deeper.
			*lines = append(*lines, fieldInfo{
				Line: fmt.Sprintf("%s%s:", indentation, fieldName),
				Help: helpText,
			})
			parseStructure(field.Type, v.Field(i), indent+1, lines)

		case reflect.Slice:
			// For slices, we append the slice name and then handle struct slices vs. primitive slices.
			*lines = append(*lines, fieldInfo{
				Line: fmt.Sprintf("%s%s:", indentation, fieldName),
				Help: helpText,
			})

			// If the slice element is another struct, we recurse into it using a zero value placeholder.
			if field.Type.Elem().Kind() == reflect.Struct {
				*lines = append(*lines, fieldInfo{
					Line: fmt.Sprintf("%s  -", indentation),
					Help: "",
				})
				parseStructure(field.Type.Elem(), reflect.Zero(field.Type.Elem()), indent+2, lines)
			} else {
				// For slices of primitives, we try to split the default value by commas.
				if defaultValue != "" {
					defaultItems := strings.Split(defaultValue, ",")
					for _, item := range defaultItems {
						item = strings.TrimSpace(item)
						*lines = append(*lines, fieldInfo{
							Line: fmt.Sprintf("%s  - %s", indentation, item),
							Help: "",
						})
					}
				} else {
					// If no default is set, provide a sample item.
					*lines = append(*lines, fieldInfo{
						Line: fmt.Sprintf("%s  - example", indentation),
						Help: "",
					})
				}
			}

		case reflect.Map:
			// For maps, we just show a sample key and value.
			*lines = append(*lines, fieldInfo{
				Line: fmt.Sprintf("%s%s:", indentation, fieldName),
				Help: helpText,
			})
			*lines = append(*lines, fieldInfo{
				Line: fmt.Sprintf("%s  key: value", indentation),
				Help: "Map example",
			})

		default:
			// For primitive fields, we assign the default or "null" if none is provided.
			value := defaultValue
			if value == "" {
				value = "null"
			} else if field.Type.Kind() == reflect.String {
				// If the field is a string, we enclose the value in quotes.
				value = fmt.Sprintf(`"%s"`, value)
			}

			*lines = append(*lines, fieldInfo{
				Line: fmt.Sprintf("%s%s: %s", indentation, fieldName, value),
				Help: helpText,
			})
		}
	}
}

// generateYAMLWithAlignment aligns the generated YAML lines with
// optional help comments on the right side.
func generateYAMLWithAlignment(lines []fieldInfo, printDescription bool) string {
	var builder strings.Builder
	maxLength := 0

	// Determine the maximum line length (without help text)
	for _, line := range lines {
		if len(line.Line) > maxLength {
			maxLength = len(line.Line)
		}
	}

	// Write lines with alignment
	for _, line := range lines {
		builder.WriteString(line.Line)
		if printDescription && line.Help != "" {
			spaces := strings.Repeat(" ", maxLength-len(line.Line)+1)
			builder.WriteString(spaces + "# " + line.Help)
		}
		builder.WriteString("\n")
	}

	return builder.String()
}

// getFieldName determines the field name to be used in YAML and for Viper lookup.
// Priority:
// 1. yaml:"..." tag (excluding "-")
// 2. mapstructure:"..." tag (excluding "-")
// 3. fallback to lowercase struct field name.
func getFieldName(field reflect.StructField) string {
	yamlName := field.Tag.Get("yaml")
	if yamlName != "" && yamlName != "-" {
		return strings.Split(yamlName, ",")[0]
	}

	mapstructureName := field.Tag.Get("mapstructure")
	if mapstructureName != "" && mapstructureName != "-" {
		return mapstructureName
	}

	return strings.ToLower(field.Name)
}

// getDefaultValue extracts the default value from struct tags.
// It first checks the "default" tag, then falls back to "placeholder".
func getDefaultValue(tag reflect.StructTag) string {
	defaultVal := tag.Get("default")
	return defaultVal
}

// getHelpText retrieves help (description) text from struct tags.
func getHelpText(tag reflect.StructTag) string {
	return tag.Get("help")
}
