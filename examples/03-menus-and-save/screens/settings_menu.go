package screens

import (
	"fmt"
	"os"

	"github.com/AveryLucas/gogogd/classdb/Button"
	"github.com/AveryLucas/gogogd/classdb/CheckBox"
	"github.com/AveryLucas/gogogd/classdb/Control"
	"github.com/AveryLucas/gogogd/classdb/Engine"
	"github.com/AveryLucas/gogogd/classdb/HSlider"
	"github.com/AveryLucas/gogogd/classdb/OptionButton"
	"github.com/AveryLucas/gogogd/gd"
	"github.com/AveryLucas/gogogd/i18n"
	"github.com/AveryLucas/gogogd/settings"
	"github.com/AveryLucas/gogogd/strict"
	"github.com/AveryLucas/gogogd/ui"

	"menus-and-save/prefs"
)

// SettingsMenu binds typed prefs.Settings to volume sliders,
// fullscreen/vsync toggles, and locale dropdown. Every widget change
// flows through Settings.Mutate → Apply.
type SettingsMenu struct {
	Control.Extension[SettingsMenu] `gd:"SettingsMenu"`

	Master     HSlider.Instance      `gd:"Panel/Grid/Master"     ggd:"strict"`
	Music      HSlider.Instance      `gd:"Panel/Grid/Music"      ggd:"strict"`
	SFX        HSlider.Instance      `gd:"Panel/Grid/SFX"        ggd:"strict"`
	Fullscreen CheckBox.Instance     `gd:"Panel/Grid/Fullscreen" ggd:"strict"`
	VSync      CheckBox.Instance     `gd:"Panel/Grid/VSync"      ggd:"strict"`
	Locale     OptionButton.Instance `gd:"Panel/Grid/Locale"     ggd:"strict"`
	Back       Button.Instance       `gd:"Panel/Back"            ggd:"strict"`
}

func (s *SettingsMenu) Ready() {
	strict.Assert(s)
	fmt.Fprintln(os.Stderr, "[03] SettingsMenu.Ready")

	cfg := settings.For[prefs.Settings]()
	cur := cfg.Get()

	s.Master.SetValue(cur.MasterVolume)
	s.Music.SetValue(cur.MusicVolume)
	s.SFX.SetValue(cur.SFXVolume)
	s.Fullscreen.SetButtonPressed(cur.Fullscreen)
	s.VSync.SetButtonPressed(cur.VSync)

	s.Locale.AddItem("English")
	s.Locale.AddItem("日本語")
	if cur.Locale == "ja" {
		s.Locale.Select(1)
	} else {
		s.Locale.Select(0)
	}

	s.Back.SetText("Back")

	apply := func(mutate func(*prefs.Settings)) {
		cfg.Mutate(mutate)
		cfg.Get().Apply()
	}

	s.Master.OnValueChanged(func(v gd.Delta) {
		apply(func(c *prefs.Settings) { c.MasterVolume = v })
	})
	s.Music.OnValueChanged(func(v gd.Delta) {
		apply(func(c *prefs.Settings) { c.MusicVolume = v })
	})
	s.SFX.OnValueChanged(func(v gd.Delta) {
		apply(func(c *prefs.Settings) { c.SFXVolume = v })
	})
	s.Fullscreen.OnToggled(func(on bool) {
		apply(func(c *prefs.Settings) { c.Fullscreen = on })
	})
	s.VSync.OnToggled(func(on bool) {
		apply(func(c *prefs.Settings) { c.VSync = on })
	})
	s.Locale.OnItemSelected(func(idx int) {
		locale := "en"
		if idx == 1 {
			locale = "ja"
		}
		apply(func(c *prefs.Settings) { c.Locale = locale })
		i18n.SetLocale(locale)
	})

	s.Back.OnPressed(func() {
		if err := cfg.Save(); err != nil {
			Engine.RaiseWarning("settings save failed:", err)
		}
		ui.Toast(s.AsNode(), "Settings saved", 1.5)
		ui.Pop(s.AsNode())
	})

	ui.SetFocus(s.Master.AsNode())
}
