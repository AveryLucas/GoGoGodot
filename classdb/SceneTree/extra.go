package SceneTree

import (
	"github.com/AveryLucas/gogogd/classdb/Engine"
	"github.com/AveryLucas/gogogd/classdb/Node"
	"github.com/AveryLucas/gogogd/classdb/Window"
	"github.com/AveryLucas/gogogd/variant/Object"
)

// Add the given node to the scene tree.
func Add(node Node.Any) {
	if tree, ok := Object.As[Instance](Engine.GetMainLoop()); ok {
		if root := tree.Root(); root != Window.Nil {
			root.AsNode().AddChild(node.AsNode())
		}
	}
}

// AddNamed adds the given node to the scene tree with the given name.
func AddNamed(name string, node Node.Any) {
	if tree, ok := Object.As[Instance](Engine.GetMainLoop()); ok {
		node := node.AsNode()
		node.SetName(name)
		if root := tree.Root(); root != Window.Nil {
			root.AsNode().AddChild(node)
		}
	}
}
