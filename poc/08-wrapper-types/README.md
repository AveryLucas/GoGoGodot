# POC #8 — Wrapper-Type Method Forwarding

The whole §6 design depends on a `Node2DExt[T]` shim that embeds `Node2D.Extension[T]` and adds method-forwarding shorthands. This POC tests whether classdb.Register copes with the extra struct layer and whether the user-facing forwarding methods actually work.

## Run

```pwsh
cd poc\08-wrapper-types
$env:CC = 'C:\Users\hello\gd\bin\zig.exe cc'
$env:CGO_ENABLED = '1'
go build -buildmode=c-shared -o graphics\windows_amd64.dll .
cd graphics
C:\Users\hello\gd\bin\godot.exe --headless
```

Full writeup: [docs/poc/08.md](../../docs/poc/08.md).
