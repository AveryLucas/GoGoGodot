// Package prefs owns the user-settings schema. Lives in its own
// package so any screen can import it without going through `main`.
package prefs

import (
	"github.com/AveryLucas/gogogd/audio"
	"github.com/AveryLucas/gogogd/gd"
	"github.com/AveryLucas/gogogd/i18n"
	"github.com/AveryLucas/gogogd/settings"
	"github.com/AveryLucas/gogogd/window"
)

// Settings is the on-disk user-preferences schema. JSON-encoded; lives
// at `user://settings_Settings.json` (the file name comes from the
// struct type name — keep it stable across releases).
type Settings struct {
	MasterVolume gd.Delta `json:"master_volume"`
	MusicVolume  gd.Delta `json:"music_volume"`
	SFXVolume    gd.Delta `json:"sfx_volume"`
	Fullscreen   bool     `json:"fullscreen"`
	VSync        bool     `json:"vsync"`
	Locale       string   `json:"locale"`
}

// Defaults returns a fresh Settings with sensible factory values.
func Defaults() Settings {
	return Settings{
		MasterVolume: 1.0,
		MusicVolume:  0.8,
		SFXVolume:    1.0,
		Fullscreen:   false,
		VSync:        true,
		Locale:       "en",
	}
}

// Apply pushes the current settings into the live engine: audio bus
// volumes, window fullscreen flag, vsync mode.
func (s Settings) Apply() {
	audio.Bus("Master").SetVolumeLinear(s.MasterVolume)
	audio.Bus("Music").SetVolumeLinear(s.MusicVolume)
	audio.Bus("SFX").SetVolumeLinear(s.SFXVolume)
	window.SetFullscreen(s.Fullscreen)
	window.SetVSync(s.VSync)
}

// Init bootstraps the typed settings singleton: applies defaults if no
// file exists, otherwise loads from disk. Call once after the engine
// is up (e.g. from MainMenu.Ready).
func Init() {
	cfg := settings.For[Settings]()
	cfg.Set(Defaults()) // baseline; Load overrides if a file exists
	_ = cfg.Load()
	cfg.Get().Apply()
	i18n.SetLocale(cfg.Get().Locale)
}
