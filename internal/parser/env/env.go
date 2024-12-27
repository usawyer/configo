package env

import (
	"encoding/json"
	"reflect"
	"strings"
)

// EnvInfo holds information needed to document an environment variable:
//   - EnvVar:       the name of the environment variable.
//   - DefaultValue: the default value (if any).
//   - HelpText:     description/help for the variable.
type EnvInfo struct {
	EnvVar       string
	DefaultValue string
	HelpText     string
	BindKey      string
	ValueType    string
}

func GetEnvs(cfg interface{}) []EnvInfo {
	var lines []EnvInfo
	parseEnvStructure(reflect.TypeOf(cfg), "", "", &lines)
	return lines
}

// parseEnvStructure recursively scans the given type (and nested structs, if any),
// collecting environment variable information according to the specified rules.
// parentPrefix will be prepended to child env tags if the parent has an env tag.
// For instance, if the parent struct has env:"db" and the nested field is env:"host",
// the final environment variable becomes "DB_HOST".
func parseEnvStructure(t reflect.Type, parentEnvPrefix, parentBindKey string, lines *[]EnvInfo) {
	// If the type is a pointer, unwrap it to its element type.
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		// Determine if the field is allowed to have an env
		envName, envAllowed := getEnvName(field)
		if !envAllowed {
			continue
		}

		msKey := getMapstructureKey(field)

		// Build the full environment variable name
		// parentEnvPrefix + "_" + envNamePart (if both are non-empty)
		childEnvName := parentEnvPrefix
		if childEnvName != "" && envName != "" {
			childEnvName += "_" + envName
		} else if envName != "" {
			childEnvName = envName
		}

		// Build the full bind key
		// parentBindKey + "." + msKey (if both are non-empty)
		childBindKey := parentBindKey
		if childBindKey != "" && msKey != "" {
			childBindKey += "." + msKey
		} else if msKey != "" {
			childBindKey = msKey
		}

		// Check the field kind to handle nested structs, slices, maps, etc.
		fieldKind := field.Type.Kind()

		// Recurse deeper if it's a struct (and not a map or slice).
		// We assume *non*-map, non-slice struct fields can have nested env variables.
		if fieldKind == reflect.Struct {
			// Recurse into nested struct.
			parseEnvStructure(field.Type, childEnvName, childBindKey, lines)
			continue
		}

		// Prepare the EnvInfo record.
		info := EnvInfo{
			EnvVar:    strings.ToUpper(childEnvName),
			BindKey:   childBindKey,
			HelpText:  getHelpText(field.Tag),
			ValueType: field.Type.String(), // e.g. "int", "[]string", "map[string]int"
		}

		// Figure out the default value. If none is provided, handle special cases for map/slice.
		defaultValStr := getDefaultValue(field.Tag)

		switch fieldKind {

		// ======================= SLICE CASE =======================
		case reflect.Slice:
			if defaultValStr == "" {
				// If no default, produce a "zero" JSON.
				elemKind := field.Type.Elem().Kind()
				if elemKind == reflect.Struct {
					// e.g. `[ {} ]`
					// Create a zero-value element, then put it into an array of length 1.
					zeroElem := reflect.Zero(field.Type.Elem()).Interface()
					oneElemArray := []interface{}{zeroElem}
					jsonBytes, _ := json.Marshal(oneElemArray)
					defaultValStr = string(jsonBytes)
				} else {
					// For slice of primitives => `[]`
					defaultValStr = "[]"
				}
			} else {
				// We assume user-supplied defaults are valid JSON.
			}
			info.DefaultValue = defaultValStr

		// ======================= MAP CASE ==========================
		case reflect.Map:
			if defaultValStr == "" {
				// No default => produce `{"key":"value"}`
				defaultValStr = `{"key":"value"}`
			}
			info.DefaultValue = defaultValStr

		// ====================== PRIMITIVE CASE =====================
		default:
			// For basic types (string, int, bool, etc.), just use whatever default is in the tag.
			info.DefaultValue = defaultValStr
		}

		*lines = append(*lines, info)
	}
}

// getEnvName determines how to name the environment variable.
// Priority:
// 1. env:"..." tag (excluding "-")
// 2. mapstructure:"..." tag => uppercase
// 3. field name => uppercase
func getEnvName(field reflect.StructField) (envName string, isAllowEnv bool) {
	defer func() {
		envName = strings.ToUpper(envName)
	}()
	// 1) Check `env` tag
	envName = field.Tag.Get("env")

	if envName == "-" {
		return "", false
	}
	if envName != "" {
		return envName, true
	}

	// 2) Fallback to mapstructure in uppercase
	msName := field.Tag.Get("mapstructure")
	if msName == "-" {
		return "", false
	}
	if msName != "" {
		return msName, true
	}

	return field.Name, true
}

// getMapstructureKey returns the part of the key used for Viper bind keys
// based on mapstructure or the field name, but does not uppercase it.
// We want something like `db` or `host`, so the final key might be `db.host`.
func getMapstructureKey(field reflect.StructField) string {
	msVal := field.Tag.Get("mapstructure")
	if msVal == "" {
		// fallback to the lowercase field name
		return strings.ToLower(field.Name)
	}
	return msVal
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
