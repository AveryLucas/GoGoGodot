/*
add_child(child_node)
await get_tree().process_frame
ensure_control_visible(child_node)
*/

package main

import (
	"github.com/AveryLucas/gogogd/classdb/SceneTree"
	"github.com/AveryLucas/gogogd/classdb/ScrollContainer"
)

var scrollContainer ScrollContainer.Instance

func ScrollContainer_EnsureControlVisible() {
	scrollContainer.AsNode().AddChild(control.AsNode())
	SceneTree.Get(scrollContainer.AsNode()).OnPhysicsFrame(func() {
		scrollContainer.EnsureControlVisible(control)
	})
}
