# POC #3 — GDScript ↔ Go Method Calls + Marshaling

## What's tested

A `Probe` Go component exposes exported methods covering int, string, `Vector2.XY`, `Float.X`, `[]int`, `bool`, and void return. A sibling `Caller` GDScript node calls each from `_ready()` and prints both the returned value and `has_method(...)` checks.

## Run

```pwsh
cd poc\03-gdscript-calls
$env:CC = 'C:\Users\hello\gd\bin\zig.exe cc'
$env:CGO_ENABLED = '1'
go build -buildmode=c-shared -o graphics\windows_amd64.dll .
cd graphics
C:\Users\hello\gd\bin\godot.exe --headless
```

Full writeup: [docs/poc/03.md](../../docs/poc/03.md).
