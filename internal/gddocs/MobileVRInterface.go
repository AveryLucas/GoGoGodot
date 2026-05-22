/*
var interface = XRServer.find_interface("Native mobile")
if interface and interface.initialize():
	get_viewport().use_xr = true
*/

package main

import (
	"github.com/AveryLucas/gogogd/classdb/Node"
	"github.com/AveryLucas/gogogd/classdb/Viewport"
	"github.com/AveryLucas/gogogd/classdb/XRInterface"
	"github.com/AveryLucas/gogogd/classdb/XRServer"
)

func ExampleMobileVR(node Node.Instance) {
	var XR = XRServer.FindInterface("Native mobile")
	if XR != XRInterface.Nil && XR.Initialize() {
		Viewport.Get(node).SetUseXr(true)
	}
}
