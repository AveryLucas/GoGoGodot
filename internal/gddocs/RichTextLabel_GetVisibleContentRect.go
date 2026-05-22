/*
[gdscript]
extends RichTextLabel

@export var background_panel: Panel

func _ready():
	await draw
	background_panel.position = get_visible_content_rect().position
	background_panel.size = get_visible_content_rect().size
[/gdscript]
[csharp]
public partial class TestLabel : RichTextLabel
{
	[Export]
	public Panel BackgroundPanel { get; set; }

	public override async void _Ready()
	{
		await ToSignal(this, Control.SignalName.Draw);
		BackgroundGPanel.Position = GetVisibleContentRect().Position;
		BackgroundPanel.Size = GetVisibleContentRect().Size;
	}
}
[/csharp]
*/

package main

import (
	"github.com/AveryLucas/gogogd/classdb/Panel"
	"github.com/AveryLucas/gogogd/classdb/RichTextLabel"
	"github.com/AveryLucas/gogogd/variant/Vector2"
)

type MyRichTextLabel struct {
	RichTextLabel.Extension[MyRichTextLabel]

	BackgroundPanel Panel.Instance
}

func (r *MyRichTextLabel) Ready() {
	r.BackgroundPanel.AsControl().SetPosition(Vector2.From(r.AsRichTextLabel().GetVisibleContentRect().Position))
	r.BackgroundPanel.AsControl().SetSize(Vector2.From(r.AsRichTextLabel().GetVisibleContentRect().Size))
}
