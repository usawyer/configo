package helper

import (
	"fmt"
	"github.com/vsysa/configo/internal/parser"
	"strings"
)

func GenerateYAMLFromTree(node *parser.ConfigNode, depth int, printDescription bool) (string, error) {
	var result strings.Builder
	indent := strings.Repeat("    ", depth) // Отступ в 4 пробела

	isRootNode := node.Level == 0

	if !isRootNode {
		if node.ConfigDescription != nil { // Если у узла есть конфигурация
			valueStr, err := formatValue(node.ConfigDescription.Default.Value)
			if err != nil {
				return "", err
			}
			comment := ""
			if printDescription && node.Description != "" {
				comment = fmt.Sprintf("  # %s", node.Description)
			}
			result.WriteString(fmt.Sprintf("%s%s: %s%s\n", indent, node.FieldName, valueStr, comment))
		} else if len(node.Children) > 0 { // Если это промежуточный узел
			// Перенос строки и комментарии для верхнего уровня
			if depth == 0 && printDescription && node.Description != "" {
				result.WriteString(fmt.Sprintf("\n# %s\n", node.Description))
			}
			result.WriteString(fmt.Sprintf("%s%s:\n", indent, node.FieldName))
		}
	}

	// Рекурсивная обработка дочерних элементов
	for _, child := range node.Children {
		nextDepth := depth + 1
		if isRootNode {
			nextDepth = 0
		}
		subTreeYaml, err := GenerateYAMLFromTree(child, nextDepth, printDescription)
		if err != nil {
			return "", err
		}
		result.WriteString(subTreeYaml)
	}

	return result.String(), nil
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
	case []string:
		var result strings.Builder
		result.WriteString("[")
		for i, elem := range v {
			if i > 0 {
				result.WriteString(", ")
			}
			result.WriteString(fmt.Sprintf("\"%s\"", elem)) // строковые элементы массива в кавычках
		}
		result.WriteString("]")
		return result.String(), nil
	case []interface{}:
		return formatSlice(v) // форматирует массив интерфейсов
	default:
		return "", fmt.Errorf("unsupported type: %T", value)
	}
}

func formatSlice(slice []interface{}) (string, error) {
	var result strings.Builder
	result.WriteString("[")
	for i, elem := range slice {
		if i > 0 {
			result.WriteString(", ")
		}
		elemStr, err := formatValue(elem)
		if err != nil {
			return "", err
		}
		result.WriteString(elemStr)
	}
	result.WriteString("]")
	return result.String(), nil
}
