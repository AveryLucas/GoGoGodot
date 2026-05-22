// Package scenetree provides convenience wrappers for the common
// SceneTree operations: Quit, ChangeScene, ReloadScene, SetPaused.
// All of these would otherwise require fishing the SceneTree out via
// `SceneTree.Get(node).Foo()`.
//
//	scenetree.ChangeScene(g.AsNode(), "res://levels/level2.tscn")
//	scenetree.SetPaused(g.AsNode(), true)
//
// For tree-walking lookups (Find, Children, Descendants, AncestorOf),
// see package `tree`.
package scenetree

import (
	"graphics.gd/classdb/Node"
	"graphics.gd/classdb/SceneTree"
)

// Quit ends the running scene tree, shutting down the engine cleanly.
//
// Requires any in-tree node to reach the active SceneTree. Quit is
// asynchronous: the engine completes the current frame before shutting
// down.
func Quit(any Node.Instance) {
	SceneTree.Get(any).Quit()
}

// ChangeScene loads the .tscn at path and replaces the current scene
// with it. Returns an error if the path is invalid or loading fails.
//
//	scenetree.ChangeScene(g.AsNode(), "res://levels/level2.tscn")
//
// Asynchronous — the change happens at the start of the next idle frame.
func ChangeScene(any Node.Instance, path string) error {
	return SceneTree.Get(any).ChangeSceneToFile(path)
}

// ReloadScene re-loads the currently-running scene from disk. Useful
// for debugging and for "restart level" patterns.
func ReloadScene(any Node.Instance) error {
	return SceneTree.Get(any).ReloadCurrentScene()
}

// SetPaused toggles whether the scene tree is processing. While paused,
// nodes with process_mode = PROCESS_MODE_ALWAYS continue to receive
// callbacks; others are suspended. Use this for pause menus.
func SetPaused(any Node.Instance, paused bool) {
	SceneTree.Get(any).SetPaused(paused)
}

// IsPaused reports whether the scene tree is currently paused.
func IsPaused(any Node.Instance) bool {
	return SceneTree.Get(any).Paused()
}
