package parser

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func ParseConfigStruct(configStruct interface{}) (*ConfigNode, error) {
	rootNode := NewRootNode()
	t := reflect.TypeOf(configStruct)
	if t.Kind() != reflect.Struct {
		return rootNode, fmt.Errorf("expected a struct type but got %s", t.Kind())
	}
	err := parseNode(rootNode, t)
	if err != nil {
		return nil, fmt.Errorf("parsing config struct error: %w", err)
	}

	return rootNode, nil
}

func parseNode(parentNode *ConfigNode, t reflect.Type) error {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldName, isHasFieldName := field.Tag.Lookup("mapstructure")
		if !isHasFieldName {
			return fmt.Errorf("field " + field.Name + " has no 'mapstructure' tag")
		}
		descTag := field.Tag.Get("desc")

		currentNode := NewConfigNode(fieldName, descTag)

		if field.Type.Kind() == reflect.Struct {
			err := parseNode(currentNode, field.Type)
			if err != nil {
				return err
			}
		} else {
			err := parseDescription(currentNode, field)
			if err != nil {
				return err
			}
		}
		err := parentNode.AddChildNode(currentNode)
		if err != nil {
			return err
		}
	}
	return nil
}

func parseDescription(configNode *ConfigNode, field reflect.StructField) error {
	if field.Type.Kind() == reflect.Struct {
		return fmt.Errorf("expected not struct type but got %s", field.Type.Kind())
	}

	defaultTag, isHasDefaultTag := field.Tag.Lookup("default")
	envTag, _ := field.Tag.Lookup("env")

	var defaultValue interface{}
	if isHasDefaultTag {
		switch field.Type.Kind() {
		case reflect.Slice:
			if field.Type.Elem().Kind() == reflect.String {
				defaultValue = strings.Split(defaultTag, ",")
			}
		case reflect.String:
			defaultValue = defaultTag
		case reflect.Int, reflect.Int32, reflect.Int64:
			intValue, err := strconv.ParseInt(defaultTag, 10, 64)
			if err != nil {
				return fmt.Errorf("cannot parse '%s' as integer", defaultTag)
			}
			defaultValue = intValue
		case reflect.Bool:
			boolValue, err := strconv.ParseBool(defaultTag)
			if err != nil {
				return fmt.Errorf("cannot parse '%s' as boolean", defaultTag)
			}
			defaultValue = boolValue
		case reflect.Float32, reflect.Float64:
			floatValue, err := strconv.ParseFloat(defaultTag, 64)
			if err != nil {
				return fmt.Errorf("cannot parse '%s' as float", defaultTag)
			}
			defaultValue = floatValue
		default:
			defaultValue = defaultTag
		}
	}

	err := configNode.SetConfigDescription(field.Type.Kind(), isHasDefaultTag, defaultValue, envTag)
	if err != nil {
		return err
	}
	return nil
}
