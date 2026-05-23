package main

import (
	"fmt"
	"math"
	"os"

	"github.com/AveryLucas/gogogd/classdb/CharacterBody2D"
	"github.com/AveryLucas/gogogd/classdb/Node"
	"github.com/AveryLucas/gogogd/gd"
	"github.com/AveryLucas/gogogd/tree"
)

// Player is a CharacterBody2D that walks toward the nearest Coin each
// physics frame. For an interactive game, swap the auto-path in
// Process for actions.Vector("left","right","up","down") — the
// surrounding code is unchanged.
type Player struct {
	CharacterBody2D.Extension[Player] `gd:"Player"`

	Speed gd.Delta
}

func (p *Player) Ready() {
	if p.Speed == 0 {
		p.Speed = 400 // pixels/sec — fast enough that the headless test
		// converges in well under a second per coin.
	}
	fmt.Fprintf(os.Stderr, "[coin-collector] Player.Ready pos=%v speed=%v\n",
		p.Position(), p.Speed)
}

func (p *Player) Process(dt gd.Delta) {
	// Find the nearest surviving coin.
	game := p.GetParent()
	if game == (Node.Instance{}) {
		return
	}
	coins := tree.Children[*Coin](game)
	if len(coins) == 0 {
		return
	}

	pos := p.GlobalPosition()
	var target *Coin
	best := gd.Delta(math.MaxFloat32)
	for _, c := range coins {
		d := distance(pos, c.GlobalPosition())
		if d < best {
			best = d
			target = c
		}
	}
	if target == nil {
		return
	}

	dir := direction(pos, target.GlobalPosition())
	p.SetVelocity(gd.Vec2{X: dir.X * p.Speed, Y: dir.Y * p.Speed})
	p.MoveAndSlide()
}

func distance(a, b gd.Vec2) gd.Delta {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return gd.Delta(math.Sqrt(float64(dx*dx + dy*dy)))
}

func direction(from, to gd.Vec2) gd.Vec2 {
	dx := to.X - from.X
	dy := to.Y - from.Y
	mag := gd.Delta(math.Sqrt(float64(dx*dx + dy*dy)))
	if mag == 0 {
		return gd.Vec2{}
	}
	return gd.Vec2{X: dx / mag, Y: dy / mag}
}
