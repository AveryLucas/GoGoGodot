// Package ui provides a modal screen stack, transient toasts, and a
// focus helper. The screen stack is a lazy CanvasLayer that lives as a
// child of the current scene root; pushed scenes become children of
// the layer.
//
//	ui.Push(g.AsNode(), "res://settings_menu.tscn")
//
//	// inside the menu's "Back" handler:
//	ui.Pop(s.AsNode())
//
// For scene-to-scene replacement (main menu → game) use
// `scenetree.ChangeScene` instead; Push doesn't replace the running
// scene, it lays over it.
package ui

import (
	"graphics.gd/classdb/CanvasLayer"
	"graphics.gd/classdb/Control"
	"graphics.gd/classdb/Label"
	"graphics.gd/classdb/Node"
	"graphics.gd/classdb/PackedScene"
	"graphics.gd/classdb/Resource"
	"graphics.gd/classdb/SceneTree"
	"graphics.gd/gd"
	"graphics.gd/timing"
	"graphics.gd/variant/Object"
)

// Push instantiates the scene at path and adds it to the modal stack as
// the topmost layer. Returns the spawned node; zero if loading fails.
//
// The first argument is any in-tree node — used to reach the active
// SceneTree and locate (or create) the UI stack layer.
func Push(in Node.Instance, scenePath string) Node.Instance {
	layer := stackLayer(in)
	if layer == (Node.Instance{}) {
		return Node.Instance{}
	}
	scene := Resource.Load[PackedScene.Instance](scenePath)
	if any(scene) == nil {
		return Node.Instance{}
	}
	node := scene.Instantiate()
	layer.AddChild(node)
	// Re-fetch: the local `node` is invalidated by AddChild's ownership
	// transfer to Godot.
	count := layer.GetChildCount()
	return layer.GetChild(count - 1)
}

// PushNode adds an already-constructed node to the modal stack. Same as
// [Push] but takes a Go value rather than a scene path — use for
// screens defined in Go code without a corresponding .tscn.
func PushNode(in Node.Instance, screen interface{ AsNode() Node.Instance }) Node.Instance {
	layer := stackLayer(in)
	if layer == (Node.Instance{}) {
		return Node.Instance{}
	}
	n := screen.AsNode()
	layer.AddChild(n)
	count := layer.GetChildCount()
	return layer.GetChild(count - 1)
}

// Pop removes the topmost modal screen and frees it. No-op if the stack
// is empty. Returns true if something was popped.
func Pop(in Node.Instance) bool {
	layer := stackLayer(in)
	if layer == (Node.Instance{}) {
		return false
	}
	count := layer.GetChildCount()
	if count == 0 {
		return false
	}
	top := layer.GetChild(count - 1)
	top.QueueFree()
	return true
}

// Clear pops every screen on the stack. Used when transitioning to a
// completely different scene (e.g. returning to main menu) so leftover
// overlays don't survive into the next scene.
func Clear(in Node.Instance) {
	layer := stackLayer(in)
	if layer == (Node.Instance{}) {
		return
	}
	count := layer.GetChildCount()
	for i := count - 1; i >= 0; i-- {
		layer.GetChild(i).QueueFree()
	}
}

// Replace pops the topmost screen and pushes scenePath in one step.
func Replace(in Node.Instance, scenePath string) Node.Instance {
	Pop(in)
	return Push(in, scenePath)
}

// Toast shows a transient on-screen message that auto-dismisses after
// duration seconds. Renders a basic Label inside the UI stack layer.
//
//	ui.Toast(g.AsNode(), "Settings saved", 1.5)
//
// Multiple toasts stack vertically (last toast on top); each manages
// its own auto-dismiss timer.
func Toast(in Node.Instance, text string, duration gd.Delta) {
	layer := stackLayer(in)
	if layer == (Node.Instance{}) {
		return
	}
	label := Label.New()
	label.SetText(text)
	label.AsControl().SetPosition(gd.Vec2{X: 16, Y: 16 + gd.Delta(layer.GetChildCount())*32})
	// Unique per-toast name so the deferred free can find this specific
	// toast even if other toasts spawn before it expires. Can't capture
	// the Label.Instance in the closure — POC #6 closure-capture trap;
	// also POC equivalent: AddChild invalidates the local handle. Look
	// up by name at fire time instead.
	name := unique("_gogogd_toast")
	label.AsNode().SetName(name)
	layer.AddChild(label.AsNode())

	timing.After(in, duration, func() {
		// `layer` is still valid (it was already in the tree before
		// this call). `label` is stale post-AddChild — look up the
		// toast by name.
		if found := layer.MoreArgs().FindChild(name, false, false); found != (Node.Instance{}) {
			found.QueueFree()
		}
	})
}

// SetFocus grabs keyboard/gamepad focus for the given Control. The node
// must already be in the tree — Godot silently ignores otherwise.
//
//	ui.SetFocus(menu.NewGame.AsNode())
func SetFocus(n Node.Instance) {
	if c, ok := Object.As[Control.Instance](n); ok {
		c.GrabFocus()
	}
}

// stackLayer returns the singleton CanvasLayer holding modal screens,
// creating it on first call. Lives as a child of the active scene root.
//
// Greppable name: `_gogogd_ui_stack`.
func stackLayer(in Node.Instance) Node.Instance {
	if in == (Node.Instance{}) {
		return Node.Instance{}
	}
	scene := SceneTree.Get(in).CurrentScene()
	if scene == (Node.Instance{}) {
		return Node.Instance{}
	}
	// Use non-owned FindChild: runtime-added children have no owner, so
	// the default FindChild (owned=true) misses them.
	if existing := scene.MoreArgs().FindChild("_gogogd_ui_stack", false, false); existing != (Node.Instance{}) {
		return existing
	}
	layer := CanvasLayer.New()
	layer.SetLayer(100)
	layer.AsNode().SetName("_gogogd_ui_stack")
	scene.AddChild(layer.AsNode())
	return layer.AsNode()
}

// unique returns a per-process-unique name with the given prefix.
var nameCounter uint64

func unique(prefix string) string {
	nameCounter++
	return prefix + "_" + itoa(nameCounter)
}

func itoa(n uint64) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
