# 01 — Coin Collector

The smallest "real game" shape gogogd makes pleasant to write. One scene, four component types, ~150 lines of Go.

**Headless validation game.** The player auto-paths toward the nearest surviving coin each physics frame, body-overlap signals fire, score increments, all coins get collected, the game exits. Replace `player.go`'s auto-path with `gogogd.InputVector("left","right","up","down")` to turn it into an interactive game; nothing else changes.

## Scene tree

```
Game (Node)                       ← Game component (tracks score, connects coins)
├── Player (CharacterBody2D)      ← Player component (auto-paths to nearest coin)
│   └── PlayerShape (CollisionShape2D)
├── Coin1 (Area2D)                ← Coin component (emits Collected on overlap)
│   └── Coin1Shape (CollisionShape2D)
├── Coin2 (Area2D)
│   └── Coin2Shape
├── Coin3 (Area2D)
│   └── Coin3Shape
└── HUD (Label)                   ← HUD component (displays score)
```

All four custom node types — `Game`, `Player`, `Coin`, `HUD` — are gogogd-registered structs. The shape resources are `CircleShape2D` sub-resources defined inline in the .tscn.

## Files

- [main.go](main.go) — one-line `main()` calling `gogogd.Run()`. Registrations are emitted by codegen into `gogogd_register.go`.
- [game.go](game.go) — the scene-root `Game`. Walks children to find coins, connects each `Coin.Collected` via `gogogd.Connect0`, updates the HUD on each pickup, calls `gogogd.Quit` when none remain.
- [player.go](player.go) — `CharacterBody2DExt[Player]` that finds the nearest coin and walks toward it via `SetVelocity` + `MoveAndSlide`.
- [coin.go](coin.go) — `Area2DExt[Coin]` with a `Collected gogogd.Signal0` field. Subscribes to Godot's built-in `body_entered` via the wrapper's `OnBodyEntered`.
- [hud.go](hud.go) — `LabelExt[HUD]` that calls `SetText("Score: N")`.

## Concepts introduced

- **Wrapper-type embedding** — `gogogd.CharacterBody2DExt[Player]`, `gogogd.Area2DExt[Coin]`, `gogogd.LabelExt[HUD]`. Method forwarders mean `p.MoveAndSlide()` and `h.SetText(...)` work without `As*()` chains.
- **Lifecycle methods** — `Ready()` runs once after the node enters the tree (children populated); `Process(dt)` runs every idle frame with `dt` typed as `gogogd.Delta` (= `float32`).
- **Custom signals** — `Coin.Collected gogogd.Signal0` is a zero-arg signal that GDScript could connect to as `coin.collected`. Emit with `c.Collected.Emit()`.
- **Owner-bound signal connections** — `gogogd.Connect0(game.AsNode(), &coin.Collected, fn)` auto-disconnects when `game` exits the tree. See [POC #7](../../docs/poc/07.md) for why this is necessary rather than stylistic.
- **Typed scene-tree walks** — `gogogd.Children[*Coin](game.AsNode())` returns only nodes that successfully cast to `*Coin`. Built on Graphics.GD's `Object.As` (see [POC #3](../../docs/poc/03.md)).
- **Vector math** — `gogogd.Vec2`, `gogogd.Delta` types, simple inline geometry helpers in `player.go`.

## Running

From this directory:

```sh
go mod tidy
../../gogogd.exe run --headless
```

Expected output:

```
[coin-collector] Player.Ready pos={0 0} speed=400
[coin-collector] Game.Ready
[coin-collector] found 3 coin(s) in scene
[coin-collector] Coin at {200 0} overlapped — emitting Collected
[coin-collector] coin collected at {200 0} — score=1
[coin-collector] Coin at {0 200} overlapped — emitting Collected
[coin-collector] coin collected at {0 200} — score=2
[coin-collector] Coin at {-200 100} overlapped — emitting Collected
[coin-collector] coin collected at {-200 100} — score=3
[coin-collector] all coins collected — exiting
```

The player visits each coin in nearest-first order (200,0 is closest to (0,0), then 0,200, then -200,100). At 400 px/sec and 60 FPS, the whole run takes roughly half a second of simulated time.

## Making it interactive

Replace `player.go`'s `Process` body with:

```go
func (p *Player) Process(dt gogogd.Delta) {
    dir := gogogd.InputVector("ui_left", "ui_right", "ui_up", "ui_down")
    p.SetVelocity(gogogd.Vec2{X: dir.X * p.Speed, Y: dir.Y * p.Speed})
    p.MoveAndSlide()
}
```

…and remove the `gogogd.Quit` call in `game.go`'s `onCoinCollected`. Then add some sprite textures and `gogogd run` (no `--headless`).
