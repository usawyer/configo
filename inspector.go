package configo

import (
	"fmt"
	"os"

	"github.com/vsysa/configo/internal/helper"
	"github.com/vsysa/configo/internal/parser"
)

// ConfigInspector предоставляет методы для анализа конфигурационной структуры.
type ConfigInspector[T Configurable] struct {
	configTree *parser.ConfigNode
}

// NewConfigInspector создает новый экземпляр ConfigInspector на основе типа конфигурации T.
func NewConfigInspector[T Configurable]() (*ConfigInspector[T], error) {
	var cfg T
	configTree, err := parser.ParseConfigStruct(cfg)
	if err != nil {
		return nil, fmt.Errorf("error parsing config struct: %w", err)
	}
	return &ConfigInspector[T]{configTree: configTree}, nil
}

// PrintConfigTemplate выводит шаблон конфигурационного файла.
func (ci *ConfigInspector[T]) PrintConfigTemplate(printWithDescription bool) {
	yamlOutput, err := helper.GenerateYAMLFromTree(ci.configTree, "", printWithDescription)
	if err != nil {
		fmt.Println("Ошибка при создании YAML:", err)
		return
	}
	fmt.Println(yamlOutput)
}

// PrintEnvHelp выводит справочную информацию о переменных окружения.
func (ci *ConfigInspector[T]) PrintEnvHelp() {
	maxLen := 0
	envCount := 0
	for _, node := range ci.configTree.GetAllLeaves() {
		envName, envExist := node.GetEnv()
		if !envExist {
			continue
		}
		l := len(envName)
		if l > maxLen {
			envCount++
			maxLen = l
		}
	}
	if envCount == 0 {
		fmt.Fprintln(os.Stdout, "No environment variables defined.")
		return
	}

	fmt.Fprintln(os.Stdout, "Environment Variables:")
	for _, node := range ci.configTree.GetAllLeaves() {
		if envName, envExist := node.GetEnv(); envExist {
			config := node.ConfigDescription
			defaultValue := ""
			if config.Default.IsExist {
				defaultValue = fmt.Sprintf("(default: %v)", config.Default.Value)
			}
			fmt.Fprintf(os.Stdout, "\t%-*s  %-10s  %s %s\n", maxLen, envName, config.ValueType.String(), node.Description, defaultValue)
		}
	}
}
