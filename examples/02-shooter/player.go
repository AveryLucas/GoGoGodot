package main

import (
	"fmt"
	"os"

	"github.com/AveryLucas/gogogd/actions"
	"github.com/AveryLucas/gogogd/classdb/Area2D"
	"github.com/AveryLucas/gogogd/classdb/CharacterBody2D"
	"github.com/AveryLucas/gogogd/classdb/Node"
	"github.com/AveryLucas/gogogd/fsm"
	"github.com/AveryLucas/gogogd/gd"
	"github.com/AveryLucas/gogogd/pool"
	"github.com/AveryLucas/gogogd/signals"
	"github.com/AveryLucas/gogogd/stat"
	"github.com/AveryLucas/gogogd/strict"
	"github.com/AveryLucas/gogogd/timing"
	"github.com/AveryLucas/gogogd/tree"
	"github.com/AveryLucas/gogogd/variant/Object"
	"github.com/AveryLucas/gogogd/visual"
)

// Player states for the FSM.
const (
	stIdle = "idle"
	stWalk = "walk"
	stHurt = "hurt"
)

// Player drives WASD movement, mouse-aim shooting, and HP. The
// CollisionShape is wired from the scene (PlayerShape); HurtBox is an
// Area2D child that fires when an Enemy enters.
type Player struct {
	CharacterBody2D.Extension[Player] `gd:"Player"`

	MaxHP    int
	Speed    gd.Delta
	FireRate gd.Delta // seconds between shots

	HurtBox Node.Instance `gd:"HurtBox" ggd:"strict"`

	HP   *stat.Stat[int]     `gd:"-"`
	Hurt signals.Signal[int] // emitted with new HP value
	Died signals.Signal0

	mode    *fsm.Machine        `gd:"-"`
	fire    timing.Cooldown     `gd:"-"`
	bullets *pool.Pool[*Bullet] `gd:"-"`
}

func NewPlayer() *Player {
	return &Player{MaxHP: 100, Speed: 260, FireRate: 0.18}
}

func (p *Player) Ready() {
	strict.Assert(p)
	p.HP = stat.New(p.MaxHP)
	visual.AttachVisualCircle(p.AsNode(), 12, visual.X11.DodgerBlue)
	fmt.Fprintf(os.Stderr, "[shooter] Player.Ready hp=%d speed=%v\n", p.HP.Get(), p.Speed)

	// HurtBox is a stock Area2D in the scene. We connect to its
	// body_entered via the typed-filter helper: anything that isn't
	// an *Enemy is dropped.
	if hb, ok := Object.As[Area2D.Instance](p.HurtBox); ok {
		hb.OnBodyEntered(tree.OnlyIfBody2D[*Enemy](func(_ *Enemy) {
			p.takeDamage(10)
		}))
	}

	// 64 bullets is plenty for fire-rate 0.18 = ~5.5/sec * ~1.5s TTL
	// = ~9 live.
	p.bullets = pool.New(p.AsNode(), 64, NewBullet)

	p.mode = fsm.New(p.AsNode(), stIdle).
		On(stIdle).
		Update(func(dt fsm.Delta) {
			if actions.Vector("ui_left", "ui_right", "ui_up", "ui_down") != (gd.Vec2{}) {
				p.mode.Goto(stWalk)
			}
		}).
		On(stWalk).
		Update(func(dt fsm.Delta) {
			if actions.Vector("ui_left", "ui_right", "ui_up", "ui_down") == (gd.Vec2{}) {
				p.mode.Goto(stIdle)
			}
		}).
		On(stHurt).
		Enter(func() {}).
		Done()
}

func (p *Player) Process(dt gd.Delta) {
	p.mode.Tick(fsm.Delta(dt))
	p.fire.Tick(dt)

	if actions.JustPressed("fire") {
		fmt.Fprintln(os.Stderr, "[shooter] fire JUST pressed")
	}
	if actions.Pressed("fire") && p.fire.Ready() {
		fmt.Fprintln(os.Stderr, "[shooter] shooting")
		p.shoot()
		p.fire.Reset(p.FireRate)
	}
}

func (p *Player) PhysicsProcess(dt gd.Delta) {
	move := actions.Vector("ui_left", "ui_right", "ui_up", "ui_down")
	p.SetVelocity(gd.Vec2{X: move.X * p.Speed, Y: move.Y * p.Speed})
	p.MoveAndSlide()
}

func (p *Player) shoot() {
	from := p.GlobalPosition()
	aim := actions.MouseDirectionFrom(from, p.AsNode())
	if aim == (gd.Vec2{}) {
		return
	}

	parent := p.GetParent()
	if parent == (Node.Instance{}) {
		return
	}

	b := p.bullets.Acquire(parent, from)
	if b == nil {
		// Pool exhausted — caller can decide. For this example, skip
		// silently; in production a debug.Watch on
		// p.bullets.Available() would catch it.
		return
	}
	b.Direction = aim
	b.OwnerFaction = FactionPlayer
	b.owner = p.bullets
	b.startTTL()
}

func (p *Player) takeDamage(dmg int) {
	hp := p.HP.Get() - dmg
	if hp < 0 {
		hp = 0
	}
	p.HP.Set(hp)
	p.Hurt.Emit(hp)
	fmt.Fprintf(os.Stderr, "[shooter] Player hurt — hp=%d\n", hp)
	if hp <= 0 {
		p.Died.Emit()
		p.QueueFree()
	}
}
