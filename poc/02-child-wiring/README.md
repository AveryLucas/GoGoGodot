# POC #2 — Child Wiring + `ggd:"strict"` Hook Point

## Hypothesis

1. Exported Node-typed fields on a registered class are auto-populated before `Ready()`.
2. `gd:"name"` controls the node name (or path) Graphics.GD looks up under the parent.
3. If the target doesn't exist in the scene, Graphics.GD silently auto-creates it.
4. We can find an observable post-Ready signal to distinguish "auto-created" from "scene-authored" so `ggd:"strict"` is implementable.

## What's tested

A `Probe` registered with five fields covering each scenario:

| Field | Tag | Scene-authored? | Expected |
|---|---|---|---|
| `AutoBare` | (none) | no | auto-create as "AutoBare" |
| `Renamed` | `gd:"CustomName"` | no | auto-create — see finding below |
| `Existing` | (none) | yes (`Existing`) | wired |
| `Nested` | `gd:"Container/Leaf"` | yes (`Container/Leaf`) | wired |
| `Skipped` | `gd:"-"` | n/a | skipped, field stays zero |

A scene with `Existing`, `Container`, and `Container/Leaf` pre-authored.

## Run

```pwsh
cd poc\02-child-wiring
$env:CC = 'C:\Users\hello\gd\bin\zig.exe cc'
$env:CGO_ENABLED = '1'
go build -buildmode=c-shared -o graphics\windows_amd64.dll .
cd graphics
C:\Users\hello\gd\bin\godot.exe --headless
```

Full writeup: [docs/poc/02.md](../../docs/poc/02.md).
