package main

import (
	"fmt"
	"os"

	"github.com/AveryLucas/gogogd/classdb/Label"
	"github.com/AveryLucas/gogogd/classdb/Node2D"
	"github.com/AveryLucas/gogogd/classdb/ProgressBar"
	"github.com/AveryLucas/gogogd/gd"
	"github.com/AveryLucas/gogogd/scenetree"
	"github.com/AveryLucas/gogogd/signals"
	"github.com/AveryLucas/gogogd/strict"
	"github.com/AveryLucas/gogogd/timing"
)

// Game is the scene root. It owns the score, wires player/spawner
// signals, and reloads the scene when the player dies.
type Game struct {
	Node2D.Extension[Game] `gd:"Game"`

	Player  *Player             `gd:"Player"    ggd:"strict"`
	Spawner *EnemySpawner       `gd:"Spawner"   ggd:"strict"`
	HPBar   ProgressBar.Instance `gd:"HUD/HP"    ggd:"strict"`
	ScoreLb Label.Instance       `gd:"HUD/Score" ggd:"strict"`

	score int
}

func (g *Game) Ready() {
	strict.Assert(g)
	fmt.Fprintln(os.Stderr, "[shooter] Game.Ready")

	g.refreshScore()

	// Bind Player.HP → HPBar. Done here (not Player.Ready) because
	// Godot runs Ready bottom-up — Player.Ready fires before
	// Game.Ready, so Game's scene-authored fields aren't populated
	// yet at Player.Ready time. By the time Game.Ready runs, both
	// Player and HPBar are valid.
	bar := g.HPBar.AsRange()
	bar.SetMaxValue(gd.Delta(g.Player.MaxHP))
	bar.SetValue(gd.Delta(g.Player.HP.Get()))
	g.Player.HP.Subscribe(func(v int) { bar.SetValue(gd.Delta(v)) })

	// Player.Hurt is emitted with the new HP value. The HUD's
	// ProgressBar reads it via Stat.Subscribe — so the only wiring
	// left at this level is "when the player dies, restart."
	signals.Connect0(g.AsNode(), &g.Player.Died, g.onPlayerDied)

	// Spawner.Spawned hands us each new enemy as it's created; we
	// chain to the enemy's Died signal to increment score. The
	// connection lives as long as Game does.
	signals.Connect(g.AsNode(), &g.Spawner.Spawned, func(e *Enemy) {
		signals.Connect0(g.AsNode(), &e.Died, func() {
			g.score++
			g.refreshScore()
		})
	})
}

func (g *Game) onPlayerDied() {
	fmt.Fprintln(os.Stderr, "[shooter] Player died — reloading scene in 1s")
	timing.After(g.AsNode(), 1.0, func() {
		_ = scenetree.ReloadScene(g.AsNode())
	})
}

func (g *Game) refreshScore() {
	g.ScoreLb.SetText(fmt.Sprintf("Kills: %d", g.score))
}
