# POC #7 — Signal owner-bound auto-disconnect on `ExitTree`

Three sibling nodes — `Emitter` (Go), `GoListener` (Go), `GdListener` (GDScript) — connect to `Emitter.tick`. At frame 3 both listeners `queue_free` themselves. We watch what happens to subsequent emits.

## Run

```pwsh
cd poc\07-signal-disconnect
$env:CC = 'C:\Users\hello\gd\bin\zig.exe cc'
$env:CGO_ENABLED = '1'
go build -buildmode=c-shared -o graphics\windows_amd64.dll .
cd graphics
C:\Users\hello\gd\bin\godot.exe --headless
```

Full writeup: [docs/poc/07.md](../../docs/poc/07.md).
