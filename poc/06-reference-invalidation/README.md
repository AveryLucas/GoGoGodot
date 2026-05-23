# POC #6 — Reference Invalidation Stress (R10)

Spawns 5 orphan `Node` instances, stashes the references in 5 different shapes, and checks `Object.InstanceIsValid` every frame for 12 frames to see which survive.

## Cases

| | Storage | Expected |
|---|---|---|
| A | unexported field on `Extension[T]` | pinned (walker reaches it) |
| B | unexported `[]Node.Instance` slice on `Extension[T]` | pinned (walker recurses) |
| C | package-level `var` | NOT pinned |
| D | closure capture (closure held on struct field) | NOT pinned (walker can't see captures) |
| E | manual `Object.Use(handle)` every frame | pinned (manual keepalive) |

## Run

```pwsh
cd poc\06-reference-invalidation
$env:CC = 'C:\Users\hello\gd\bin\zig.exe cc'
$env:CGO_ENABLED = '1'
go build -buildmode=c-shared -o graphics\windows_amd64.dll .
cd graphics
C:\Users\hello\gd\bin\godot.exe --headless
```

Full writeup: [docs/poc/06.md](../../docs/poc/06.md).
