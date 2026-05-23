// 03-menus-and-save — vertical slice exercising the production-systems
// surface: typed save slots with migrations, typed settings with
// reactive UI, modal screen stack, i18n with re-translation, audio /
// window control.
//
// Flow:
//
//	MainMenu  ──→  New Game ──→  Game ──→  PauseMenu ──→  (Resume / Save / ToMenu)
//	   │                                       │
//	   ├─→ Continue (loads autosave) ──────────↑
//	   ├─→ Settings ──→ SettingsMenu (overlay)
//	   └─→ Quit
//
// Run: `gogogd run` (or `gogogd test --duration 5s` for headless).
package main

import (
	"github.com/AveryLucas/gogogd/startup"

	// Blank-import every subpackage that owns registered types so each
	// subpackage's gogogd_register.go init() runs at startup.
	_ "menus-and-save/game"
	_ "menus-and-save/prefs"
	_ "menus-and-save/screens"
)

func main() {
	startup.Scene()
}
