package helper

import (
	"fmt"
	"strings"

	"github.com/vsysa/configo/internal/parser"
)

func GenerateYAMLFromTree(node *parser.ConfigNode, indent string, printDescription bool) (string, error) {
	var result strings.Builder

	isRootNode := node.Level == 0

	// Обработка узлов не корневого уровня
	if !isRootNode {
		// Массив структур
		if node.IsArrayOfStructs {
			if printDescription && node.Description != "" {
				result.WriteString(fmt.Sprintf("%s# %s\n", indent, node.Description))
			}
			result.WriteString(fmt.Sprintf("%s%s:\n", indent, node.FieldName))
			// Генерация массива структур
			indentOfStruct := indent + "    "
			dash := "-"
			for _, childNode := range node.Children {
				itemYAML, err := GenerateYAMLFromTree(childNode, indentOfStruct+"  ", printDescription)
				if err != nil {
					return "", err
				}
				result.WriteString(fmt.Sprintf("%s%s %s\n", indentOfStruct, dash, strings.TrimSpace(itemYAML)))
				dash = " "
			}
		} else if node.ConfigDescription != nil {
			if node.ConfigDescription.IsArray {
				// Массив примитивов
				if printDescription && node.Description != "" {
					result.WriteString(fmt.Sprintf("%s# %s\n", indent, node.Description))
				}
				result.WriteString(fmt.Sprintf("%s%s:\n", indent, node.FieldName))

				// Приведение к []interface{}
				var slice []interface{}
				switch v := node.ConfigDescription.Default.Value.(type) {
				case []string:
					for _, elem := range v {
						slice = append(slice, elem)
					}
				case []int:
					for _, elem := range v {
						slice = append(slice, elem)
					}
				case []float64:
					for _, elem := range v {
						slice = append(slice, elem)
					}
				case []bool:
					for _, elem := range v {
						slice = append(slice, elem)
					}
				case []interface{}:
					slice = v
				default:
					return "", fmt.Errorf("unsupported slice type: %T", node.ConfigDescription.Default.Value)
				}

				// Генерация элементов массива
				for _, elem := range slice {
					elemStr, err := formatValue(elem)
					if err != nil {
						return "", err
					}
					result.WriteString(fmt.Sprintf("%s- %s\n", indent+"  ", elemStr))
				}
			} else {
				// Примитивные типы
				valueStr, err := formatValue(node.ConfigDescription.Default.Value)
				if err != nil {
					return "", err
				}
				comment := ""
				if printDescription && node.Description != "" {
					comment = fmt.Sprintf("  # %s", node.Description)
				}
				result.WriteString(fmt.Sprintf("%s%s: %s%s\n", indent, node.FieldName, valueStr, comment))
			}
		} else if len(node.Children) > 0 {
			// Если это промежуточный узел
			if printDescription && node.Description != "" {
				result.WriteString(fmt.Sprintf("%s# %s\n", indent, node.Description))
			}
			result.WriteString(fmt.Sprintf("%s%s:\n", indent, node.FieldName))
		}
	}

	// Рекурсивная обработка дочерних элементов
	if !node.IsArrayOfStructs {
		for _, child := range node.Children {
			nextIndent := indent + "    "
			if isRootNode {
				nextIndent = ""
			}
			subTreeYaml, err := GenerateYAMLFromTree(child, nextIndent, printDescription)
			if err != nil {
				return "", err
			}
			result.WriteString(subTreeYaml)
		}
	}

	return result.String(), nil
}

func calculateCommentIndent(line string, maxWidth int) string {
	lineLength := len(line)
	if lineLength >= maxWidth {
		return " "
	}
	return strings.Repeat(" ", maxWidth-lineLength)
}

func formatValue(value interface{}) (string, error) {
	if value == nil {
		return "null", nil // YAML-стиль для nil значений
	}
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("\"%s\"", v), nil
	case bool:
		return fmt.Sprintf("%v", v), nil // возвращает "true" или "false" без кавычек
	case int, int32, int64, uint, uint32, uint64, float32, float64:
		return fmt.Sprintf("%v", v), nil // возвращает числа без кавычек
	default:
		return "", fmt.Errorf("unsupported type: %T", value)
	}
}
