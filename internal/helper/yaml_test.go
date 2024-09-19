package helper

import (
	"github.com/stretchr/testify/assert"
	"github.com/vsysa/configo/internal/parser"
	"testing"
)

// Простое дерево с одним узлом
func TestGenerateYAMLFromTree_SingleNode(t *testing.T) {
	node := &parser.ConfigNode{
		FieldName:   "app_name",
		Level:       1,
		Description: "Application name",
		ConfigDescription: &parser.ConfigDescription{
			Default: struct {
				IsExist bool
				Value   interface{}
			}{Value: "TestApp", IsExist: true},
		},
		Children: nil,
	}

	expectedYAML := "app_name: \"TestApp\"\n"

	result, err := GenerateYAMLFromTree(node, 0, false)
	assert.NoError(t, err)
	assert.Equal(t, expectedYAML, result)
}

// Дерево с вложенными узлами
func TestGenerateYAMLFromTree_NestedNodes(t *testing.T) {
	node := &parser.ConfigNode{
		FieldName:   "database",
		Level:       1,
		Description: "Database configuration",
		Children: []*parser.ConfigNode{
			{
				FieldName: "host",
				Level:     2,
				ConfigDescription: &parser.ConfigDescription{
					Default: struct {
						IsExist bool
						Value   interface{}
					}{Value: "localhost", IsExist: true},
				},
				Description: "Database host",
			},
			{
				FieldName: "port",
				Level:     2,
				ConfigDescription: &parser.ConfigDescription{
					Default: struct {
						IsExist bool
						Value   interface{}
					}{Value: 5432, IsExist: true},
				},
				Description: "Database port",
			},
		},
	}

	expectedYAML := `database:
    host: "localhost"
    port: 5432
`

	result, err := GenerateYAMLFromTree(node, 0, false)
	assert.NoError(t, err)
	assert.Equal(t, expectedYAML, result)
}

// Дерево с массивами и булевыми значениями
func TestGenerateYAMLFromTree_WithArraysAndBooleans(t *testing.T) {
	node := &parser.ConfigNode{
		FieldName:   "config",
		Level:       1,
		Description: "Sample configuration",
		Children: []*parser.ConfigNode{
			{
				FieldName: "features",
				Level:     2,
				ConfigDescription: &parser.ConfigDescription{
					Default: struct {
						IsExist bool
						Value   interface{}
					}{Value: []string{"feature1", "feature2"}, IsExist: true},
				},
				Description: "Enabled features",
			},
			{
				FieldName: "debug",
				Level:     2,
				ConfigDescription: &parser.ConfigDescription{
					Default: struct {
						IsExist bool
						Value   interface{}
					}{Value: true, IsExist: true},
				},
				Description: "Debug mode",
			},
		},
	}

	expectedYAML := `config:
    features: ["feature1", "feature2"]
    debug: true
`

	result, err := GenerateYAMLFromTree(node, 0, false)
	assert.NoError(t, err)
	assert.Equal(t, expectedYAML, result)
}

// Проверка генерации комментариев при printDescription = true
func TestGenerateYAMLFromTree_WithDescriptions(t *testing.T) {
	node := &parser.ConfigNode{
		FieldName:   "server",
		Level:       1,
		Description: "Server configuration",
		Children: []*parser.ConfigNode{
			{
				FieldName: "host",
				Level:     2,
				ConfigDescription: &parser.ConfigDescription{
					Default: struct {
						IsExist bool
						Value   interface{}
					}{Value: "0.0.0.0", IsExist: true},
				},
				Description: "Server host",
			},
			{
				FieldName: "port",
				Level:     2,
				ConfigDescription: &parser.ConfigDescription{
					Default: struct {
						IsExist bool
						Value   interface{}
					}{Value: 80, IsExist: true},
				},
				Description: "Server port",
			},
		},
	}

	expectedYAML := `
# Server configuration
server:
    host: "0.0.0.0"  # Server host
    port: 80  # Server port
`

	result, err := GenerateYAMLFromTree(node, 0, true)
	assert.NoError(t, err)
	assert.Equal(t, expectedYAML, result)
}
