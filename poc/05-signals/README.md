# POC #5 — Signal Declaration + Go ↔ GDScript Round-Trip

A `Probe` declares three signals as struct fields (`Void`, `Solo[int]`, `Pair[string, Vector2.XY]`), Go connects local handlers in `Ready`, GDScript connects via `signal.connect()` in `_ready`, and `Probe.Process` emits each once.

## Run

```pwsh
cd poc\05-signals
$env:CC = 'C:\Users\hello\gd\bin\zig.exe cc'
$env:CGO_ENABLED = '1'
go build -buildmode=c-shared -o graphics\windows_amd64.dll .
cd graphics
C:\Users\hello\gd\bin\godot.exe --headless
```

Full writeup: [docs/poc/05.md](../../docs/poc/05.md).
