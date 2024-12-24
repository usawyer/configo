package helper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vsysa/configo/internal/parser"
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

	result, err := GenerateYAMLFromTree(node, "", false)
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

	result, err := GenerateYAMLFromTree(node, "", false)
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
					IsArray: true,
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
    features:
      - "feature1"
      - "feature2"
    debug: true
`

	result, err := GenerateYAMLFromTree(node, "", false)
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

	expectedYAML := `# Server configuration
server:
    host: "0.0.0.0"  # Server host
    port: 80  # Server port
`

	result, err := GenerateYAMLFromTree(node, "", true)
	assert.NoError(t, err)
	assert.Equal(t, expectedYAML, result)
}

func TestGenerateYAMLFromTree_WithNestedStructAndPrimitiveArrays(t *testing.T) {
	node := &parser.ConfigNode{
		FieldName:   "config",
		Level:       1,
		Description: "Complex configuration with nested arrays",
		Children: []*parser.ConfigNode{
			{
				FieldName:        "devices",
				Level:            2,
				IsArrayOfStructs: true,
				Description:      "List of devices",
				Children: []*parser.ConfigNode{
					{
						FieldName: "host",
						Level:     3,
						ConfigDescription: &parser.ConfigDescription{
							Default: struct {
								IsExist bool
								Value   interface{}
							}{Value: "127.0.0.1", IsExist: true},
						},
						Description: "Device host",
					},
					{
						FieldName: "port",
						Level:     3,
						ConfigDescription: &parser.ConfigDescription{
							Default: struct {
								IsExist bool
								Value   interface{}
							}{Value: 8080, IsExist: true},
						},
						Description: "Device port",
					},
					{
						FieldName:        "settings",
						Level:            3,
						IsArrayOfStructs: true,
						Description:      "Device settings",
						Children: []*parser.ConfigNode{
							{
								FieldName: "param",
								Level:     4,
								ConfigDescription: &parser.ConfigDescription{
									Default: struct {
										IsExist bool
										Value   interface{}
									}{Value: "default", IsExist: true},
								},
								Description: "Parameter name",
							},
							{
								FieldName: "value",
								Level:     4,
								ConfigDescription: &parser.ConfigDescription{
									Default: struct {
										IsExist bool
										Value   interface{}
									}{Value: 42, IsExist: true},
								},
								Description: "Parameter value",
							},
						},
					},
					{
						FieldName:   "ports",
						Level:       3,
						Description: "List of ports for the device",
						ConfigDescription: &parser.ConfigDescription{
							IsArray: true,
							Default: struct {
								IsExist bool
								Value   interface{}
							}{Value: []int{80, 443, 9090}, IsExist: true},
						},
					},
				},
			},
			{
				FieldName: "names",
				Level:     2,
				ConfigDescription: &parser.ConfigDescription{
					IsArray: true,
					Default: struct {
						IsExist bool
						Value   interface{}
					}{Value: []string{"Alice", "Bob"}, IsExist: true},
				},
				Description: "List of names",
			},
		},
	}

	expectedYAML := `# Complex configuration with nested arrays
config:
    # List of devices
    devices:
        - host: "127.0.0.1"  # Device host
          port: 8080  # Device port
          # Device settings
          settings:
              - param: "default"  # Parameter name
                value: 42  # Parameter value
          # List of ports for the device
          ports:
            - 80
            - 443
            - 9090
    # List of names
    names:
      - "Alice"
      - "Bob"
`

	result, err := GenerateYAMLFromTree(node, "", true)
	assert.NoError(t, err)
	assert.Equal(t, expectedYAML, result)
}
