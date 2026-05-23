package game

import (
	"fmt"
	"os"

	"github.com/AveryLucas/gogogd/actions"
	"github.com/AveryLucas/gogogd/classdb/CharacterBody2D"
	"github.com/AveryLucas/gogogd/fx"
	"github.com/AveryLucas/gogogd/gd"
	"github.com/AveryLucas/gogogd/signals"
	"github.com/AveryLucas/gogogd/visual"
)

// Player walks with WASD. Visual is a blue circle. Takes damage on
// the M key (debug); dies when HP hits 0.
type Player struct {
	CharacterBody2D.Extension[Player] `gd:"Player"`

	MaxHP int
	HP    int `gd:"-"`
	Speed gd.Delta

	Died signals.Signal0
}

func NewPlayer() *Player {
	return &Player{MaxHP: 3, Speed: 220}
}

func (p *Player) Ready() {
	p.HP = p.MaxHP
	visual.AttachVisualCircle(p.AsNode(), 16, visual.X11.DodgerBlue)
	fmt.Fprintf(os.Stderr, "[03] Player.Ready hp=%d\n", p.HP)
}

func (p *Player) PhysicsProcess(dt gd.Delta) {
	move := actions.Vector("ui_left", "ui_right", "ui_up", "ui_down")
	p.SetVelocity(gd.Vec2{X: move.X * p.Speed, Y: move.Y * p.Speed})
	p.MoveAndSlide()
}

func (p *Player) Process(dt gd.Delta) {
	if actions.JustPressed("hurt_self") {
		p.takeDamage(1)
	}
}

func (p *Player) takeDamage(dmg int) {
	p.HP -= dmg
	fx.Flash(p.AsNode(), 0.12)
	fmt.Fprintf(os.Stderr, "[03] Player hurt — hp=%d\n", p.HP)
	if p.HP <= 0 {
		p.Died.Emit()
		p.QueueFree()
	}
}
