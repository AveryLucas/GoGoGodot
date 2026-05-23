package screens

import (
	"fmt"
	"os"

	"github.com/AveryLucas/gogogd/classdb/Button"
	"github.com/AveryLucas/gogogd/classdb/Control"
	"github.com/AveryLucas/gogogd/scenetree"
	"github.com/AveryLucas/gogogd/strict"
	"github.com/AveryLucas/gogogd/ui"

	"menus-and-save/game"
	"menus-and-save/prefs"
	"menus-and-save/saves"
)

// MainMenu is the title screen: New Game / Continue / Settings / Quit.
type MainMenu struct {
	Control.Extension[MainMenu] `gd:"MainMenu"`

	NewGame  Button.Instance `gd:"Panel/VBox/NewGame"  ggd:"strict"`
	Continue Button.Instance `gd:"Panel/VBox/Continue" ggd:"strict"`
	Settings Button.Instance `gd:"Panel/VBox/Settings" ggd:"strict"`
	Quit     Button.Instance `gd:"Panel/VBox/Quit"     ggd:"strict"`
}

func (m *MainMenu) Ready() {
	strict.Assert(m)
	fmt.Fprintln(os.Stderr, "[03] MainMenu.Ready")

	// Init settings on first scene Ready (engine singletons are alive
	// by now). Idempotent — calling Init twice just re-reads the file.
	prefs.Init()

	m.NewGame.SetText("New Game")
	m.Continue.SetText("Continue")
	m.Settings.SetText("Settings")
	m.Quit.SetText("Quit")

	m.Continue.SetDisabled(!saves.Exists("autosave"))

	m.NewGame.OnPressed(func() {
		_ = scenetree.ChangeScene(m.AsNode(), "res://game.tscn")
	})
	m.Continue.OnPressed(func() {
		game.SetPendingLoadSlot("autosave")
		_ = scenetree.ChangeScene(m.AsNode(), "res://game.tscn")
	})
	m.Settings.OnPressed(func() {
		ui.Push(m.AsNode(), "res://settings_menu.tscn")
	})
	m.Quit.OnPressed(func() {
		scenetree.Quit(m.AsNode())
	})

	ui.SetFocus(m.NewGame.AsNode())
}
