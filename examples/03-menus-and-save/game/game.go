package game

import (
	"fmt"
	"os"

	"github.com/AveryLucas/gogogd/bus"
	"github.com/AveryLucas/gogogd/classdb/Engine"
	"github.com/AveryLucas/gogogd/classdb/Node2D"
	"github.com/AveryLucas/gogogd/gd"
	"github.com/AveryLucas/gogogd/scenetree"
	"github.com/AveryLucas/gogogd/signals"
	"github.com/AveryLucas/gogogd/strict"
	"github.com/AveryLucas/gogogd/timing"

	"menus-and-save/saves"
)

// Game is the gameplay scene. Tiny by design — the interesting
// surface is the wrapper (menus, saves, settings), not the gameplay.
//
// Player walks with WASD. P opens the pause menu. M takes damage
// (player dies after 3 hits) → emits final_score → GameOver scene.
//
// Auto-saves to slot "autosave" every 30 seconds; manual saves from
// the pause menu also write to "autosave."
type Game struct {
	Node2D.Extension[Game] `gd:"Game"`

	Player *Player `gd:"Player" ggd:"strict"`

	Score    int    `gd:"-"`
	loadSlot string `gd:"-"`
}

func (g *Game) Ready() {
	strict.Assert(g)
	fmt.Fprintln(os.Stderr, "[03] Game.Ready")

	g.loadSlot = pendingLoadSlot
	pendingLoadSlot = ""
	if g.loadSlot != "" {
		g.loadFrom(g.loadSlot)
	}

	// Pause menu fires this via the bus.
	bus.On("save_requested").Connect(g.AsNode(), func() {
		g.saveTo("autosave")
	})

	// Player death → emit final score and transition.
	signals.Connect0(g.AsNode(), &g.Player.Died, func() {
		bus.EmitVal("final_score", g.Score)
		timing.After(g.AsNode(), 1.0, func() {
			_ = scenetree.ChangeScene(g.AsNode(), "res://game_over.tscn")
		})
	})

	// Autosave every 30 seconds.
	timing.Every(g.AsNode(), 30, func() { g.saveTo("autosave") })

	// Tick score up — gives "Save" a visible effect on later load.
	timing.Every(g.AsNode(), 1.0, func() { g.Score++ })
}

// Process is currently empty (pause menu opens via UI on key press;
// would need actions package import to bind that key). Keeping the
// stub so the lifecycle hook stays visible.
func (g *Game) Process(dt gd.Delta) {
	// pause handling would go here; deferred per scope.
}

func (g *Game) saveTo(slotName string) {
	pos := g.Player.GlobalPosition()
	if err := saves.Write(slotName, g.Score, 1, pos.X, pos.Y); err != nil {
		Engine.RaiseWarning("save failed:", err)
		return
	}
	fmt.Fprintf(os.Stderr, "[03] saved to %q (score=%d pos=%v)\n", slotName, g.Score, pos)
}

func (g *Game) loadFrom(slotName string) {
	state, err := saves.Slot(slotName).Read()
	if err != nil {
		Engine.RaiseWarning("load failed:", err)
		return
	}
	g.Score = state.Score
	g.Player.SetGlobalPosition(gd.Vec2{X: state.PlayerX, Y: state.PlayerY})
	fmt.Fprintf(os.Stderr, "[03] loaded from %q (score=%d pos=(%v,%v))\n",
		slotName, state.Score, state.PlayerX, state.PlayerY)
}

// pendingLoadSlot is the cross-scene baton for "Continue was clicked,
// load this slot." Set by MainMenu before ChangeScene; read by Game.Ready.
// Package-level var because bus.EmitVal is synchronous — by the time
// Game boots, no listeners exist to receive the slot name.
var pendingLoadSlot string

// SetPendingLoadSlot is called by MainMenu before requesting a scene
// change. Lives on the game package so screens import game/ rather
// than the reverse.
func SetPendingLoadSlot(slot string) { pendingLoadSlot = slot }
