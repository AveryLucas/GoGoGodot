package startup

import (
	gd "github.com/AveryLucas/gogogd/internal"
	"github.com/AveryLucas/gogogd/internal/gdextension"
	"github.com/AveryLucas/gogogd/internal/gdreference"
	"github.com/AveryLucas/gogogd/internal/pointers"
	"github.com/AveryLucas/gogogd/internal/ring"
	"github.com/AveryLucas/gogogd/variant/Callable"

	_ "unsafe"
)

//go:linkname keep_reachable_instances_alive github.com/AveryLucas/gogogd/classdb.keep_reachable_instances_alive
func keep_reachable_instances_alive()

func init() {
	gdextension.On.MainLoop.EveryFrame = func() {
		Callable.Cycle()
		ring.Main.Flush()
		keep_reachable_instances_alive()
		gdreference.GC(gd.Free)
		pointers.Cycle()
	}
	gdextension.On.MainLoop.FinalFrame = func() {

	}
}
