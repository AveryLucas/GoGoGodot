package gogogd

import (
	"github.com/AveryLucas/gogogd/classdb/Node"
	"github.com/AveryLucas/gogogd/scenetree"
)

// Quit shuts the engine down cleanly. `any` is any node in the tree
// (used to reach SceneTree).
func Quit(any Node.Instance) { scenetree.Quit(any) }

// ChangeScene switches to the scene at `path` (typically "res://x.tscn").
func ChangeScene(any Node.Instance, path string) error {
	return scenetree.ChangeScene(any, path)
}

// ReloadScene reloads the current scene from disk.
func ReloadScene(any Node.Instance) error { return scenetree.ReloadScene(any) }

// SetPaused toggles the engine's pause state.
func SetPaused(any Node.Instance, paused bool) { scenetree.SetPaused(any, paused) }

// IsPaused reports whether the engine is currently paused.
func IsPaused(any Node.Instance) bool { return scenetree.IsPaused(any) }
