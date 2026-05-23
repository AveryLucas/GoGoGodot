package screens

import (
	"fmt"
	"os"

	"github.com/AveryLucas/gogogd/bus"
	"github.com/AveryLucas/gogogd/classdb/Button"
	"github.com/AveryLucas/gogogd/classdb/Control"
	"github.com/AveryLucas/gogogd/scenetree"
	"github.com/AveryLucas/gogogd/strict"
	"github.com/AveryLucas/gogogd/ui"
)

// PauseMenu overlays the running game. Pauses the scene tree; resumes
// on Resume; ToMenu clears the stack and returns to main menu. Save
// fires a bus event so the Game owns the actual save.
type PauseMenu struct {
	Control.Extension[PauseMenu] `gd:"PauseMenu"`

	Resume   Button.Instance `gd:"Panel/VBox/Resume"   ggd:"strict"`
	Settings Button.Instance `gd:"Panel/VBox/Settings" ggd:"strict"`
	Save     Button.Instance `gd:"Panel/VBox/Save"     ggd:"strict"`
	ToMenu   Button.Instance `gd:"Panel/VBox/ToMenu"   ggd:"strict"`
}

func (p *PauseMenu) Ready() {
	strict.Assert(p)
	fmt.Fprintln(os.Stderr, "[03] PauseMenu.Ready — pausing scene tree")
	scenetree.SetPaused(p.AsNode(), true)

	p.Resume.SetText("Resume")
	p.Settings.SetText("Settings")
	p.Save.SetText("Save")
	p.ToMenu.SetText("Main Menu")

	p.Resume.OnPressed(func() {
		scenetree.SetPaused(p.AsNode(), false)
		ui.Pop(p.AsNode())
	})
	p.Settings.OnPressed(func() {
		ui.Push(p.AsNode(), "res://settings_menu.tscn")
	})
	p.Save.OnPressed(func() {
		bus.Emit("save_requested")
		ui.Toast(p.AsNode(), "Game saved", 1.5)
	})
	p.ToMenu.OnPressed(func() {
		scenetree.SetPaused(p.AsNode(), false)
		ui.Clear(p.AsNode())
		_ = scenetree.ChangeScene(p.AsNode(), "res://main_menu.tscn")
	})

	ui.SetFocus(p.Resume.AsNode())
}

// ExitTree clears the pause flag even if the menu was dismissed via
// scene change rather than the Resume button.
func (p *PauseMenu) ExitTree() {
	scenetree.SetPaused(p.AsNode(), false)
}
