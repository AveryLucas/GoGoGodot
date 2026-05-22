package gd

import (
	"github.com/AveryLucas/gogogd/internal/gdextension"
	"github.com/AveryLucas/gogogd/internal/gdmemory"
)

func init() {
	gdextension.On.Editor = gdextension.CallbacksForEditor{
		ClassInUseDetection: func(classes gdextension.PackedArray[gdextension.String], result gdextension.Returns[gdextension.PackedArray[gdextension.String]]) {
			gdmemory.Set(gdextension.Pointer(result), classes)
		},
	}
}
