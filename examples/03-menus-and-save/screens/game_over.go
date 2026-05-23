package screens

import (
	"fmt"
	"os"

	"github.com/AveryLucas/gogogd/bus"
	"github.com/AveryLucas/gogogd/classdb/Button"
	"github.com/AveryLucas/gogogd/classdb/Control"
	"github.com/AveryLucas/gogogd/classdb/Label"
	"github.com/AveryLucas/gogogd/scenetree"
	"github.com/AveryLucas/gogogd/strict"
	"github.com/AveryLucas/gogogd/ui"
)

// GameOver shows the final score and offers Retry / Main Menu.
type GameOver struct {
	Control.Extension[GameOver] `gd:"GameOver"`

	ScoreLb Label.Instance  `gd:"Panel/Score"  ggd:"strict"`
	Retry   Button.Instance `gd:"Panel/Retry"  ggd:"strict"`
	ToMenu  Button.Instance `gd:"Panel/ToMenu" ggd:"strict"`

	FinalScore int `gd:"-"`
}

func (g *GameOver) Ready() {
	strict.Assert(g)
	fmt.Fprintln(os.Stderr, "[03] GameOver.Ready")

	bus.ConnectVal(bus.On("final_score"), g.AsNode(), func(v int) {
		g.FinalScore = v
		g.refresh()
	})
	g.refresh()

	g.Retry.SetText("Retry")
	g.ToMenu.SetText("Main Menu")

	g.Retry.OnPressed(func() {
		_ = scenetree.ChangeScene(g.AsNode(), "res://game.tscn")
	})
	g.ToMenu.OnPressed(func() {
		_ = scenetree.ChangeScene(g.AsNode(), "res://main_menu.tscn")
	})

	ui.SetFocus(g.Retry.AsNode())
}

func (g *GameOver) refresh() {
	g.ScoreLb.SetText(fmt.Sprintf("Final Score: %d", g.FinalScore))
}
