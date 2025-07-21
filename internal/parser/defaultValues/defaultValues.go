package defaultValues

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type DefaultInfo struct {
	BindKey      string
	DefaultValue interface{}
}

func GetDefaultValues(cfg interface{}) ([]DefaultInfo, error) {
	var lines []DefaultInfo
	err := parseDefaultValues(reflect.TypeOf(cfg), "", &lines)
	return lines, err
}

func parseDefaultValues(t reflect.Type, parentBindKey string, lines *[]DefaultInfo) error {
	// If the type is a pointer, unwrap it to its element type.
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("not a struct")
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		msKey := getMapstructureKey(field)
		// Build the full bind key
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
			err := parseDefaultValues(field.Type, childBindKey, lines)
			if err != nil {
				return err
			}
			continue
		}

		defaultValStr := getDefaultValue(field.Tag)
		if defaultValStr == "" {
			continue
		}

		var defaultValue interface{}

		if fieldKind == reflect.Slice {
			if !isPrimitive(field.Type.Elem().Kind()) {
				// array of non primitives not allowed
				continue
			}
			if string(defaultValStr[0]) == "[" && string(defaultValStr[len(defaultValStr)-1]) == "]" {
				// Creating a new slice using reflect
				sliceType := reflect.SliceOf(field.Type.Elem())
				slicePtr := reflect.New(sliceType)

				// Decompressing JSON into a slice
				err := json.Unmarshal([]byte(defaultValStr), slicePtr.Interface())
				if err != nil {
					fmt.Printf("cannot unmarshal default value \"%s\" as %s: %s", defaultValStr, field.Type.String(), err)
					continue
				}
				defaultValue = slicePtr.Elem().Interface()
			} else {
				defaultValue = strings.Split(defaultValStr, ",")
			}

		} else if field.Type.Kind() == reflect.Map {
			// Creating a map type
			mapType := reflect.MapOf(field.Type.Key(), field.Type.Elem())
			mapPtr := reflect.New(mapType)

			err := json.Unmarshal([]byte(defaultValStr), mapPtr.Interface())
			if err != nil {
				fmt.Printf("cannot unmarshal default value \"%s\" as %s: %s", defaultValStr, field.Type.String(), err)
				continue
			}

			defaultValue = mapPtr.Elem().Interface()
		} else if field.Type == reflect.TypeOf(time.Duration(0)) {
			durationVal, err := time.ParseDuration(defaultValStr)
			if err != nil {
				fmt.Printf("cannot parse default value \"%s\" as duration: %s", defaultValStr, err)
				continue
			}
			defaultValue = durationVal
		} else {
			// Processing of single values (primitives)
			switch fieldKind {
			case reflect.String:
				defaultValue = defaultValStr
			case reflect.Int, reflect.Int32, reflect.Int64:
				intValue, err := strconv.ParseInt(defaultValStr, 10, 64)
				if err != nil {
					fmt.Printf("cannot parse default value '%s' as integer", defaultValStr)
					continue
				}
				defaultValue = intValue
			case reflect.Bool:
				boolValue, err := strconv.ParseBool(defaultValStr)
				if err != nil {
					fmt.Printf("cannot parse default value '%s' as boolean", defaultValStr)
					continue
				}
				defaultValue = boolValue
			case reflect.Float32, reflect.Float64:
				floatValue, err := strconv.ParseFloat(defaultValStr, 64)
				if err != nil {
					fmt.Printf("cannot parse default value '%s' as float", defaultValStr)
					continue
				}
				defaultValue = floatValue
			default:
				defaultValue = defaultValStr
			}
		}

		*lines = append(*lines, DefaultInfo{
			BindKey:      childBindKey,
			DefaultValue: defaultValue,
		})
	}
	return nil
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

func isPrimitive(kind reflect.Kind) bool {
	switch kind {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String:
		return true
	default:
		return false
	}
}
