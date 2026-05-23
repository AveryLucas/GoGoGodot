# 02 — Top-Down Shooter

A small action-game vertical slice. The acceptance test for Phase 2 of the
[implementation plan](../../docs/IMPLEMENTATION_PLAN.md).

**You play:** a top-down character with WASD movement and mouse-aim shooting.
Enemies spawn at the screen edges and chase you. Hits decrement HP via a
reactive `Stat[int]` bound to the HUD bar; dying reloads the scene.

## Run

```
cd examples/02-shooter
gogogd run            # opens a window
gogogd run --headless # for verification only — no input, but the loop ticks
```

## Scene tree

```
Game (Node2D)              ← Game component
├── Player (CharacterBody2D)  ← Player component
│   ├── PlayerShape (CollisionShape2D)
│   └── HurtBox (Area2D)
│       └── HurtBoxShape (CollisionShape2D)
├── Spawner (EnemySpawner)    ← spawns enemies at screen edges
└── HUD (CanvasLayer)
    ├── HP (ProgressBar)      ← bound to Player.HP via Stat.Subscribe
    └── Score (Label)         ← driven by Game.refreshScore
```

Bullets and enemies are Go-defined types instantiated at runtime with
`gogogd.Add(parent, NewEnemy(), pos)` and `pool.Acquire(parent, pos)`. No
`.tscn` files for them.

## Files

- `main.go` — `gogogd.Run()`. Registrations are emitted by `gogogd register`
  into `gogogd_register.go`.
- `game.go` — root component; binds Player.HP → HPBar, tracks score, reloads
  on player death.
- `player.go` — WASD + mouse-aim shooter; flat FSM with idle/walk/hurt;
  fire-rate cooldown; bullet pool.
- `bullet.go` — straight-line motion, body-enter hit detection, TTL despawn
  via pool release.
- `enemy.go` — chases player; dies on damage (no death animation — that's
  Phase 3 fx territory).
- `enemy_spawner.go` — periodic `gogogd.Every` spawner at screen edges.

## What this example exercises (Phase 2 surface)

- `gogogd.Pool[*Bullet]` — recycled-spawn pool sized for fast firing.
- `gogogd.Add` — polymorphic spawn (Go value + position).
- `gogogd.Find[T]`, `gogogd.Children[T]`, `gogogd.AncestorOf[T]` — tree queries.
- `gogogd.OnlyIfBody2D[T]` — typed body-entered signal filtering.
- `gogogd.Connect[T]` / `Signal0` / `Signal[T]` — owner-bound signal lifetime.
- `gogogd.MousePos` / `gogogd.MouseDirectionFrom` — aiming.
- `gogogd.Cooldown` — fire-rate gating.
- `gogogd.Every` / `gogogd.After` — spawn cadence and TTL-based despawn.
- `fsm.Machine` — Player state machine.
- `gogogd.Stat[int]` + `ProgressBar.SetValue` — reactive HUD binding.
- `gogogd.NodeInstance` / `gogogd.Area2DInstance` / `gogogd.LabelInstance`
  / `gogogd.ProgressBarInstance` — typed scene-wired field aliases.

## What is NOT in this example (deferred to Phase 3)

The aspirational shooter design in earlier drafts of `docs/ARCHITECTURE.md`
included these polish helpers — they belong to Phase 3 (production systems)
and are deliberately absent:

- `gogogd.Flash`, `gogogd.Hitstop`, `gogogd.Camera().Shake` — fx helpers.
- `gogogd.Particles(asset, pos)`, `gogogd.SoundAt(asset, pos)` — one-shots.
- `gogogd.PlaySound("res://...", gogogd.SFX{Volume: -6})` — audio config-struct.
- `gogogd.Transition.Fade(...).To(...)` — scene transitions.
- `gogogd.Bus` — global event bus.

Adding these to the shooter is the natural Phase 3 exit-gate followup.

## Known gaps

- The example doesn't have sprite art — characters render as outline-only
  via "Visible Collision Shapes" in the editor. Phase 2 plan acknowledged
  this in the [Phase 1 ship-state notes](../../docs/IMPLEMENTATION_PLAN.md);
  add colored rectangles when Phase 3 ships proper VFX.
- The HUD's HPBar is wired by `Game.Ready` (after `Player.Ready` populates
  `Player.HP`). The order is Godot's bottom-up Ready dispatch — children
  call Ready first. Cross-component wiring belongs in the parent.
- Player auto-attack input is keyboard-mouse only — joypad/touch input is
  Phase 2's "Input breadth" subtask, not yet wired into this example.
