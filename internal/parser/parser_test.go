package parser

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Тест для ParseConfigStruct: проверка правильного парсинга структуры с тегами.
func TestParseConfigStruct_Success(t *testing.T) {
	type SubConfig struct {
		Field1 string `mapstructure:"field1" desc:"This is field 1"`
		Field2 int    `mapstructure:"field2" desc:"This is field 2" default:"42"`
	}

	type TestConfig struct {
		SubConfig SubConfig `mapstructure:"sub_config" desc:"Sub config section"`
		BoolField bool      `mapstructure:"bool_field" desc:"A boolean field" default:"true"`
	}

	var cfg TestConfig
	rootNode, err := ParseConfigStruct(cfg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Проверка, что корневой узел содержит дочерние узлы
	if len(rootNode.Children) != 2 {
		t.Errorf("Expected 2 child nodes, got %d", len(rootNode.Children))
	}

	// Проверка, что первый дочерний узел — SubConfig
	subConfigNode := rootNode.Children[0]
	if subConfigNode.FieldName != "sub_config" {
		t.Errorf("Expected first child to be 'sub_config', got %s", subConfigNode.FieldName)
	}

	// Проверка описания для узла SubConfig
	if subConfigNode.Description != "Sub config section" {
		t.Errorf("Expected description to be 'Sub config section', got %s", subConfigNode.Description)
	}

	// Проверка дочерних узлов SubConfig
	if len(subConfigNode.Children) != 2 {
		t.Errorf("Expected 2 children for SubConfig, got %d", len(subConfigNode.Children))
	}

	// Проверка узла Field2 в SubConfig
	field2Node := subConfigNode.Children[1]
	if field2Node.FieldName != "field2" {
		t.Errorf("Expected field2 in SubConfig, got %s", field2Node.FieldName)
	}
	if field2Node.ConfigDescription.Default.Value != int64(42) {
		t.Errorf("Expected default value for field2 to be 42, got %v", field2Node.ConfigDescription.Default.Value)
	}

	// Проверка узла для BoolField
	boolFieldNode := rootNode.Children[1]
	if boolFieldNode.FieldName != "bool_field" {
		t.Errorf("Expected 'bool_field', got %s", boolFieldNode.FieldName)
	}
	if boolFieldNode.ConfigDescription.Default.Value != true {
		t.Errorf("Expected default value for bool_field to be true, got %v", boolFieldNode.ConfigDescription.Default.Value)
	}
}

// Тест на обработку ошибки, если передана не структура
func TestParseConfigStruct_InvalidType(t *testing.T) {
	var notStruct string
	_, err := ParseConfigStruct(notStruct)
	if err == nil {
		t.Fatalf("Expected error for non-struct type, but got nil")
	}
	expectedError := "expected a struct type but got string"
	if err.Error() != expectedError {
		t.Errorf("Expected error: %s, got: %s", expectedError, err.Error())
	}
}

// Тест на парсинг структуры с отсутствующим тегом "mapstructure"
func TestParseConfigStruct_MissingMapStructureTag(t *testing.T) {
	type InvalidConfig struct {
		FieldWithoutTag string `desc:"This field has no mapstructure tag"`
	}

	var cfg InvalidConfig
	_, err := ParseConfigStruct(cfg)
	if err == nil {
		t.Fatalf("Expected error for missing 'mapstructure' tag, but got nil")
	}
	expectedError := "parsing config struct error: field FieldWithoutTag has no 'mapstructure' tag"
	if err.Error() != expectedError {
		t.Errorf("Expected error: %s, got: %s", expectedError, err.Error())
	}
}

// Тест для проверки парсинга сложного типа с тегом default
func TestParseDescription(t *testing.T) {
	field := reflect.StructField{
		Name: "TestField",
		Type: reflect.TypeOf(""),
		Tag:  `mapstructure:"test_field" default:"hello" desc:"test field"`,
	}

	node := NewConfigNode("test_field", "test field")
	err := parseDescription(node, field)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if node.ConfigDescription == nil {
		t.Fatal("ConfigDescription should not be nil")
	}

	if node.ConfigDescription.Default.Value != "hello" {
		t.Errorf("Expected default value 'hello', got %v", node.ConfigDescription.Default.Value)
	}
}

func TestParseConfigStruct_ArrayOfStructs(t *testing.T) {
	type Signal struct {
		Label string `mapstructure:"label" desc:"Signal label" default:"default_label"`
		Do    int    `mapstructure:"do" desc:"Signal action" default:"1"`
	}

	type Device struct {
		Host    string   `mapstructure:"host" desc:"Device host"`
		Port    int      `mapstructure:"port" desc:"Device port"`
		Signals []Signal `mapstructure:"signals" desc:"List of signals"`
	}

	type TestConfig struct {
		Devices []Device `mapstructure:"devices" desc:"List of devices"`
	}

	var cfg TestConfig
	rootNode, err := ParseConfigStruct(cfg)
	require.NoError(t, err, "Unexpected error during parsing config")

	require.Len(t, rootNode.Children, 1, "Expected 1 child node at the root level")

	devicesNode := rootNode.Children[0]
	assert.Equal(t, "devices", devicesNode.FieldName, "Expected first child node to be 'devices'")
	assert.True(t, devicesNode.IsArrayOfStructs, "Expected 'devices' node to be marked as array")

	require.Len(t, devicesNode.Children, 3, "Expected 3 children for 'devices' (host, port, signals)")

	hostNode := devicesNode.Children[0]
	assert.Equal(t, "host", hostNode.FieldName, "Expected first child of 'devices' to be 'host'")

	signalsNode := devicesNode.Children[2]
	assert.Equal(t, "signals", signalsNode.FieldName, "Expected third child of 'devices' to be 'signals'")
	assert.True(t, signalsNode.IsArrayOfStructs, "Expected 'signals' node to be marked as array")

	require.Len(t, signalsNode.Children, 2, "Expected 2 children for 'signals' (label, do)")

	labelNode := signalsNode.Children[0]
	assert.Equal(t, "label", labelNode.FieldName, "Expected first child of 'signals' to be 'label'")
	assert.Equal(t, "default_label", labelNode.ConfigDescription.Default.Value, "Expected default value for 'label'")
}

func TestParseConfigStruct_ArrayOfPrimitives(t *testing.T) {
	type TestConfig struct {
		Names []string `mapstructure:"names" desc:"List of names" default:"Alice,Bob,Charlie"`
		Ports []int    `mapstructure:"ports" desc:"List of ports" default:"8080,9090"`
	}

	var cfg TestConfig
	rootNode, err := ParseConfigStruct(cfg)
	require.NoError(t, err, "Unexpected error during parsing config")

	require.Len(t, rootNode.Children, 2, "Expected 2 child nodes at the root level")

	// Проверка узла 'names'
	namesNode := rootNode.Children[0]
	assert.Equal(t, "names", namesNode.FieldName, "Expected first child node to be 'names'")
	assert.False(t, namesNode.IsArrayOfStructs, "Expected 'names' node to not be marked as array of structs")
	assert.True(t, namesNode.ConfigDescription.IsArray, "Expected 'names' node type to be marked as array of primitives")
	assert.Equal(t, "string", namesNode.ConfigDescription.ValueType.String(), "Expected 'names' node type to be string")

	defaultNames := namesNode.ConfigDescription.Default.Value.([]string)
	require.Len(t, defaultNames, 3, "Expected 3 default names")
	assert.Equal(t, "Alice", defaultNames[0], "First default name should be Alice")

	// Проверка узла 'ports'
	portsNode := rootNode.Children[1]
	assert.Equal(t, "ports", portsNode.FieldName, "Expected second child node to be 'ports'")
	assert.False(t, portsNode.IsArrayOfStructs, "Expected 'ports' node to not be marked as array of structs")
	assert.True(t, portsNode.ConfigDescription.IsArray, "Expected 'ports' node type to be marked as array of primitives")
	assert.Equal(t, "int", portsNode.ConfigDescription.ValueType.String(), "Expected 'ports' node type to be int")

	defaultPorts := portsNode.ConfigDescription.Default.Value.([]int)
	require.Len(t, defaultPorts, 2, "Expected 2 default ports")
	assert.Equal(t, 8080, defaultPorts[0], "First default port should be 8080")
}

func TestParseConfigStruct_EmptyArray(t *testing.T) {
	type TestConfig struct {
		EmptyArray []string `mapstructure:"empty_array" desc:"An empty array"`
	}

	var cfg TestConfig
	rootNode, err := ParseConfigStruct(cfg)
	require.NoError(t, err, "Unexpected error during parsing config")

	require.Len(t, rootNode.Children, 1, "Expected 1 child node at the root level")

	emptyArrayNode := rootNode.Children[0]
	assert.Equal(t, "empty_array", emptyArrayNode.FieldName, "Expected child node to be 'empty_array'")
	assert.False(t, emptyArrayNode.IsArrayOfStructs, "Expected 'empty_array' to not be marked as array")
	assert.True(t, emptyArrayNode.ConfigDescription.IsArray, "Expected 'empty_array' node type to be marked as array of primitives")
	assert.Equal(t, "string", emptyArrayNode.ConfigDescription.ValueType.String(), "Expected 'empty_array' node type to be string")
	assert.False(t, emptyArrayNode.ConfigDescription.Default.IsExist, "Expected no default value for 'empty_array'")
}
