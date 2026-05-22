package Startup

import (
	gd "github.com/AveryLucas/gogogd/internal"
	"github.com/AveryLucas/gogogd/internal/gdextension"
)

func init() {
	gd.LinkStartup = func() {
		sname = gdextension.Host.Strings.Intern.UTF8("GodotInstance")
		otype = gdextension.Host.Objects.Type(sname)
		gd.LinkMethods(sname, &methods, false)
	}
}
