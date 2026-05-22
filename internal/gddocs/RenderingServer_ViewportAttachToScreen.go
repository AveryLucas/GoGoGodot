/*
[gdscript]
func _ready():
	RenderingServer.viewport_attach_to_screen(get_viewport().get_viewport_rid(), Rect2())
	RenderingServer.viewport_attach_to_screen($Viewport.get_viewport_rid(), Rect2(0, 0, 600, 600))
[/gdscript]
*/

package main

import (
	"github.com/AveryLucas/gogogd/classdb/RenderingServer"
	"github.com/AveryLucas/gogogd/classdb/Viewport"
	"github.com/AveryLucas/gogogd/variant/Rect2"
)

var viewport Viewport.Instance

func RenderingServer_ViewportAttachToScreen() {
	RenderingServer.ViewportAttachToScreen(Viewport.Get(node).GetViewportRid(), Rect2.PositionSize{}, 0)
	RenderingServer.ViewportAttachToScreen(viewport.GetViewportRid(), Rect2.New(0, 0, 600, 600), 0)
}
