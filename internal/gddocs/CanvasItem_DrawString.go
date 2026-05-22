/*
[gdscript]
draw_string(ThemeDB.fallback_font, Vector2(64, 64), "Hello world", HORIZONTAL_ALIGNMENT_LEFT, -1, ThemeDB.fallback_font_size)
[/gdscript]
[csharp]
DrawString(ThemeDB.FallbackFont, new Vector2(64, 64), "Hello world", HorizontalAlignment.Left, -1, ThemeDB.FallbackFontSize);
[/csharp]
*/

package main

import (
	"github.com/AveryLucas/gogogd/classdb/CanvasItem"
	"github.com/AveryLucas/gogogd/classdb/GUI"
	"github.com/AveryLucas/gogogd/classdb/TextServer"
	"github.com/AveryLucas/gogogd/classdb/ThemeDB"
	"github.com/AveryLucas/gogogd/variant/Color"
	"github.com/AveryLucas/gogogd/variant/Vector2"
)

var canvas_item CanvasItem.Instance

func CanvasItem_DrawString() {
	canvas_item.MoreArgs().DrawString(ThemeDB.FallbackFont(), Vector2.New(64, 64), "Hello world", GUI.HorizontalAlignmentLeft, -1, ThemeDB.FallbackFontSize(),
		Color.W3C.White, TextServer.JustificationKashida|TextServer.JustificationWordBound, 0, 0, 0)
}
