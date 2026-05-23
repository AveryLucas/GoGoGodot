package main

import (
	"fmt"
	"os"

	"github.com/AveryLucas/gogogd/classdb/Node"
	"github.com/AveryLucas/gogogd/scenetree"
	"github.com/AveryLucas/gogogd/signals"
	"github.com/AveryLucas/gogogd/tree"
)

// Game is the scene root. It tracks score, updates the HUD when coins
// are picked up, and quits the engine when every coin has been
// collected.
type Game struct {
	Node.Extension[Game] `gd:"Game"`

	score int
}

func (g *Game) Ready() {
	fmt.Fprintln(os.Stderr, "[coin-collector] Game.Ready")

	// Find every coin in the scene and wire its Collected signal to
	// our score-update handler. signals.Connect0 binds the connection
	// to g's lifetime: when Game exits the tree, all of these
	// auto-disconnect.
	coins := tree.Children[*Coin](g.AsNode())
	fmt.Fprintf(os.Stderr, "[coin-collector] found %d coin(s) in scene\n", len(coins))

	for _, c := range coins {
		coin := c // capture
		signals.Connect0(g.AsNode(), &coin.Collected, func() {
			g.onCoinCollected(coin)
		})
	}
	g.updateHUD()
}

func (g *Game) onCoinCollected(c *Coin) {
	g.score++
	fmt.Fprintf(os.Stderr, "[coin-collector] coin collected at %v — score=%d\n",
		c.GlobalPosition(), g.score)
	g.updateHUD()
	c.QueueFree()

	// Recount surviving coins. The freed coin is already queued for
	// deletion but still in the tree this frame, so we filter it out.
	remaining := tree.Children[*Coin](g.AsNode())
	alive := 0
	for _, x := range remaining {
		if x != c {
			alive++
		}
	}
	if alive == 0 {
		fmt.Fprintln(os.Stderr, "[coin-collector] all coins collected — exiting")
		os.Stderr.Sync()
		scenetree.Quit(g.AsNode())
	}
}

func (g *Game) updateHUD() {
	huds := tree.Children[*HUD](g.AsNode())
	if len(huds) == 0 {
		return
	}
	huds[0].SetScore(g.score)
}
