package main

import (
	"fmt"
	"os"

	"github.com/AveryLucas/gogogd/classdb/Area2D"
	"github.com/AveryLucas/gogogd/signals"
	"github.com/AveryLucas/gogogd/tree"
)

// Coin is an Area2D that emits a Collected signal when the Player
// overlaps it. The Game (parent) connects to this signal and frees the
// coin after reacting.
type Coin struct {
	Area2D.Extension[Coin] `gd:"Coin"`

	Collected signals.Signal0
}

func (c *Coin) Ready() {
	// Subscribe to body_entered through tree.OnlyIfBody2D, so the
	// handler only fires for bodies that cast to *Player. Anything
	// else (debris, another coin, an enemy in a future example) is
	// silently dropped.
	c.OnBodyEntered(tree.OnlyIfBody2D[*Player](func(p *Player) {
		fmt.Fprintf(os.Stderr, "[coin-collector] Coin at %v collected by Player (speed=%v) — emitting Collected\n",
			c.GlobalPosition(), p.Speed)
		c.Collected.Emit()
	}))
}
