# POC #4 — Lifecycle Ordering

Verifies that **constructor → property deserialize → child wiring → `Ready`** holds in practice, and surfaces two hooks (`OnCreate`, `Init`) that ARCHITECTURE.md doesn't currently mention.

## Run

```pwsh
cd poc\04-lifecycle-order
$env:CC = 'C:\Users\hello\gd\bin\zig.exe cc'
$env:CGO_ENABLED = '1'
go build -buildmode=c-shared -o graphics\windows_amd64.dll .
cd graphics
C:\Users\hello\gd\bin\godot.exe --headless
```

Full writeup: [docs/poc/04.md](../../docs/poc/04.md).
