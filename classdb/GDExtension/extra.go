package GDExtension

import (
	gd "github.com/AveryLucas/gogogd/internal"
	"github.com/AveryLucas/gogogd/internal/gdextension"
	"github.com/AveryLucas/gogogd/internal/pointers"
)

// LibraryPath is the path to the shared library that contains the current GD extension.
func LibraryPath() string {
	return pointers.New[gd.String](gdextension.Host.Library.Location()).String()
}
