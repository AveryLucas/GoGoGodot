// Package window exposes the main-window controls that user settings
// touch: fullscreen toggle, vsync mode.
//
//	window.SetFullscreen(true)
//	window.SetVSync(false)
//
// For finer window control (resize, position, multi-monitor) drop
// down to `graphics.gd/classdb/DisplayServer` directly.
package window

import (
	"graphics.gd/classdb/DisplayServer"
)

// SetFullscreen toggles the main window's mode between windowed and
// fullscreen.
func SetFullscreen(on bool) {
	mode := DisplayServer.WindowModeWindowed
	if on {
		mode = DisplayServer.WindowModeFullscreen
	}
	DisplayServer.WindowSetMode(mode, DisplayServer.MainWindowId)
}

// IsFullscreen reports whether the main window is currently in any
// fullscreen mode (includes borderless-fullscreen).
func IsFullscreen() bool {
	mode := DisplayServer.WindowGetMode(DisplayServer.MainWindowId)
	return mode == DisplayServer.WindowModeFullscreen ||
		mode == DisplayServer.WindowModeExclusiveFullscreen
}

// SetVSync enables or disables vertical sync on the main window.
// Enabled is the standard "no tearing, capped at refresh" mode.
//
// For Adaptive vsync, use [SetVSyncMode] with DisplayServer.VsyncAdaptive.
func SetVSync(on bool) {
	mode := DisplayServer.VsyncDisabled
	if on {
		mode = DisplayServer.VsyncEnabled
	}
	DisplayServer.WindowSetVsyncMode(mode, DisplayServer.MainWindowId)
}

// SetVSyncMode sets the explicit vsync mode for users who need Adaptive
// or Mailbox modes. Pass DisplayServer.Vsync* constants.
func SetVSyncMode(mode DisplayServer.VSyncMode) {
	DisplayServer.WindowSetVsyncMode(mode, DisplayServer.MainWindowId)
}
