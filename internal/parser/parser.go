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
		envTag, _ := field.Tag.Lookup("env")

		currentNode := NewConfigNode(fieldName, descTag)
		currentNode.EnvName = envTag
		err := parentNode.AddChildNode(currentNode)
		if err != nil {
			return err
		}

		if field.Type.Kind() == reflect.Slice {
			// Обработка массивов

			if field.Type.Elem().Kind() == reflect.Struct {
				// Если элемент массива - структура, разбираем её
				currentNode.IsArrayOfStructs = true
				err := parseNode(currentNode, field.Type.Elem())
				if err != nil {
					return err
				}
			} else {
				// Если элемент массива - примитивный тип
				err := parseDescription(currentNode, field)
				if err != nil {
					return err
				}
			}
		} else if field.Type.Kind() == reflect.Struct {
			// Если это структура, рекурсивно разбираем её
			err := parseNode(currentNode, field.Type)
			if err != nil {
				return err
			}
		} else {
			// Обрабатываем обычные примитивы
			err := parseDescription(currentNode, field)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func parseDescription(configNode *ConfigNode, field reflect.StructField) error {
	if field.Type.Kind() == reflect.Struct {
		return fmt.Errorf("expected not struct type but got %s", field.Type.Kind())
	}

	defaultTag, isHasDefaultTag := field.Tag.Lookup("default")

	isArray := field.Type.Kind() == reflect.Slice
	valueType := field.Type.Kind()
	if isArray {
		valueType = field.Type.Elem().Kind()
	}

	var defaultValue interface{}
	if isHasDefaultTag {
		if isArray {
			switch valueType {
			case reflect.String:
				// Массив строк, разделённых запятыми
				defaultValue = strings.Split(defaultTag, ",")
			case reflect.Int, reflect.Int32, reflect.Int64:
				// Массив целых чисел
				strValues := strings.Split(defaultTag, ",")
				intValues := make([]int, len(strValues))
				for i, strVal := range strValues {
					intVal, err := strconv.Atoi(strVal)
					if err != nil {
						return fmt.Errorf("cannot parse '%s' as integer in array", strVal)
					}
					intValues[i] = intVal
				}
				defaultValue = intValues
			case reflect.Float32, reflect.Float64:
				// Массив чисел с плавающей точкой
				strValues := strings.Split(defaultTag, ",")
				floatValues := make([]float64, len(strValues))
				for i, strVal := range strValues {
					floatVal, err := strconv.ParseFloat(strVal, 64)
					if err != nil {
						return fmt.Errorf("cannot parse '%s' as float in array", strVal)
					}
					floatValues[i] = floatVal
				}
				defaultValue = floatValues
			case reflect.Bool:
				// Массив булевых значений
				strValues := strings.Split(defaultTag, ",")
				boolValues := make([]bool, len(strValues))
				for i, strVal := range strValues {
					boolVal, err := strconv.ParseBool(strVal)
					if err != nil {
						return fmt.Errorf("cannot parse '%s' as boolean in array", strVal)
					}
					boolValues[i] = boolVal
				}
				defaultValue = boolValues
			default:
				return fmt.Errorf("unsupported slice type: %s", valueType)
			}
		} else {
			// Обработка одиночных значений (примитивы)
			switch field.Type.Kind() {
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
	}

	err := configNode.SetConfigDescription(isArray, valueType, isHasDefaultTag, defaultValue)
	if err != nil {
		return err
	}
	return nil
}
