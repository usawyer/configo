package parser

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRootNode(t *testing.T) {
	root := NewRootNode()
	assert.Equal(t, "root", root.FieldName)
	assert.Equal(t, "root node", root.Description)
}

func TestAddChildNode(t *testing.T) {
	root := NewRootNode()
	child := NewConfigNode("child1", "description")

	err := root.AddChildNode(child)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(root.Children))
	assert.Equal(t, root, child.Parent)
}

func TestAddChildToConfigNodeWithDescription(t *testing.T) {
	root := NewRootNode()
	root.SetConfigDescription(reflect.String, true, "default")

	child := NewConfigNode("child1", "description")
	err := root.AddChildNode(child)
	assert.Error(t, err)
}

func TestSetConfigDescription(t *testing.T) {
	root := NewRootNode()
	node := NewConfigNode("child1", "description")
	err := node.SetConfigDescription(reflect.String, true, "default")
	assert.NoError(t, err)
	root.AddChildNode(node)

	assert.NotNil(t, node.ConfigDescription)
	assert.Equal(t, reflect.String, node.ConfigDescription.ValueType)
	assert.Equal(t, "", node.EnvName)
}

func TestGetFullPathParts(t *testing.T) {
	root := NewRootNode()
	child := NewConfigNode("child1", "description")
	root.AddChildNode(child)

	// Проверяем, что путь для корневого узла — это просто "root"
	expectedRootPath := []string{}
	assert.Equal(t, expectedRootPath, root.GetFullPathParts())

	// Проверяем, что для дочернего узла путь включает как "root", так и "child1"
	expectedChildPath := []string{"child1"}
	assert.Equal(t, expectedChildPath, child.GetFullPathParts())
}

func TestGetAllLeaves(t *testing.T) {
	root := NewRootNode()
	child := NewConfigNode("child1", "description")
	child.SetConfigDescription(reflect.Int, true, 10)
	root.AddChildNode(child)

	leaves := root.GetAllLeaves()
	assert.Equal(t, 1, len(leaves))
	assert.Equal(t, "child1", leaves[0].FieldName)
}
