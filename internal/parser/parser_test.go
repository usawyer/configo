package parser

import (
	"reflect"
	"testing"
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
