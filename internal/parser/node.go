package parser

import (
	"fmt"
	"reflect"
	"strings"
)

type ConfigDescription struct {
	ConfigNode *ConfigNode
	Env        struct {
		IsExist bool
		Path    string
	}
	ValueType reflect.Kind
	Default   struct {
		IsExist bool
		Value   interface{}
	}
}
type ConfigNode struct {
	FieldName   string
	Description string
	Children    []*ConfigNode
	Parent      *ConfigNode
	Level       int

	ConfigDescription *ConfigDescription
}

// GenerateEnv генерирует env на основе пути к конфигу без учета переопределения тегом env
func (r *ConfigNode) GenerateEnv() string {
	if r.Parent != nil {
		return strings.ToUpper(strings.Join(r.GetFullPathParts(), "_"))
	}
	// Если у узла нет родителя (т.е. это корневой узел), то у него нет env тк корневой всегда один - root
	return ""
}

// GetEnv отдает итоговый env с учетом переопределения через тег
func (r *ConfigNode) GetEnv() string {
	if r.ConfigDescription.Env.IsExist {
		return r.ConfigDescription.Env.Path
	}
	return r.GenerateEnv()
}

func NewRootNode() *ConfigNode {
	return &ConfigNode{FieldName: "root", Description: "root node"}
}

func NewConfigNode(fieldName string, description string) *ConfigNode {
	return &ConfigNode{FieldName: fieldName, Description: description}
}

// получение пути вместе с текущим узлом
func (r *ConfigNode) GetFullPathParts() []string {
	// Если родителя нет, возвращаем путь, состоящий только из имени текущего узла
	if r.Parent == nil {
		return []string{}
	}
	pathParts := r.Parent.GetFullPathParts()
	return append(pathParts, r.FieldName)
}

func (r *ConfigNode) GetAllNodesDeep() []*ConfigNode {
	var nodes []*ConfigNode
	for _, child := range r.Children {
		nodes = append(nodes, child)
		nodes = append(nodes, child.GetAllNodesDeep()...)
	}
	return nodes
}

func (r *ConfigNode) GetAllLeaves() []*ConfigNode {
	var items []*ConfigNode
	if r.ConfigDescription != nil {
		items = append(items, r)
	}
	for _, child := range r.Children {
		items = append(items, child.GetAllLeaves()...)
	}
	return items
}

func (r *ConfigNode) GetAllParentNodes() []*ConfigNode {
	var nodes []*ConfigNode
	for current := r.Parent; current != nil; current = current.Parent {
		nodes = append(nodes, current)
	}
	return nodes
}

func (r *ConfigNode) AddChildNode(node *ConfigNode) error {
	if r.ConfigDescription != nil {
		return fmt.Errorf("item in node != nil. adding children node is not possible. name: %s", node.FieldName)
	}
	node.Parent = r
	node.Level = r.Level + 1
	r.Children = append(r.Children, node)
	return nil
}

func (r *ConfigNode) SetConfigDescription(ValueType reflect.Kind, isDefaultExist bool, defaultValue interface{}, env string) error {
	if len(r.Children) > 0 {
		return fmt.Errorf("children in node != 0. setting item to node is not possible, node: %s", r.FieldName)
	}
	r.ConfigDescription = &ConfigDescription{
		Env: struct {
			IsExist bool
			Path    string
		}{IsExist: false, Path: ""},

		ValueType: ValueType,
		Default: struct {
			IsExist bool
			Value   interface{}
		}{
			IsExist: isDefaultExist,
			Value:   defaultValue,
		},
	}

	if len(env) > 0 {
		r.ConfigDescription.Env.IsExist = true
		r.ConfigDescription.Env.Path = strings.ToUpper(env)
	}

	return nil
}
