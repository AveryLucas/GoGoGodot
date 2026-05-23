package main

import (
	"math/rand"
	"time"

	"github.com/AveryLucas/gogogd/classdb/Node2D"
	"github.com/AveryLucas/gogogd/gd"
	"github.com/AveryLucas/gogogd/signals"
	"github.com/AveryLucas/gogogd/spawn"
	"github.com/AveryLucas/gogogd/timing"
	"github.com/AveryLucas/gogogd/tree"
)

// EnemySpawner emits a new Enemy every Interval seconds at a random
// screen-edge position. The spawned enemy targets the live Player
// (located once at Ready via tree.Find).
type EnemySpawner struct {
	Node2D.Extension[EnemySpawner] `gd:"EnemySpawner"`

	Interval gd.Delta

	Spawned signals.Signal[*Enemy]

	player *Player    `gd:"-"`
	rng    *rand.Rand `gd:"-"`
}

func NewEnemySpawner() *EnemySpawner {
	return &EnemySpawner{Interval: 1.3}
}

func (s *EnemySpawner) Ready() {
	if p, ok := tree.Find[*Player](s.AsNode()); ok {
		s.player = p
	}
	s.rng = rand.New(rand.NewSource(time.Now().UnixNano()))

	parent := s.GetParent()
	timing.Every(s.AsNode(), s.Interval, func() {
		if s.player == nil {
			return
		}
		e := NewEnemy()
		e.Target = s.player

		pos := s.randomScreenEdge(720, 480, 64)
		spawn.Add(parent, e, pos)
		s.Spawned.Emit(e)
	})
}

// randomScreenEdge picks a position margin pixels outside one of the
// four edges of the viewport-sized box (0,0)..(width,height).
func (s *EnemySpawner) randomScreenEdge(width, height, margin gd.Delta) gd.Vec2 {
	side := s.rng.Intn(4)
	switch side {
	case 0: // top
		return gd.Vec2{X: gd.Delta(s.rng.Float32()) * width, Y: -margin}
	case 1: // bottom
		return gd.Vec2{X: gd.Delta(s.rng.Float32()) * width, Y: height + margin}
	case 2: // left
		return gd.Vec2{X: -margin, Y: gd.Delta(s.rng.Float32()) * height}
	default: // right
		return gd.Vec2{X: width + margin, Y: gd.Delta(s.rng.Float32()) * height}
	}
}
