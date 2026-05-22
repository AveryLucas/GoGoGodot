/*
[gdscript]
var tween = get_tree().create_tween()
tween.tween_property($Sprite, "modulate", Color.RED, 1).set_trans(Tween.TRANS_SINE)
tween.tween_property($Sprite, "scale", Vector2(), 1).set_trans(Tween.TRANS_BOUNCE)
tween.tween_callback($Sprite.queue_free)
[/gdscript]
[csharp]
Tween tween = GetTree().CreateTween();
tween.TweenProperty(GetNode("Sprite"), "modulate", Colors.Red, 1.0f).SetTrans(Tween.TransitionType.Sine);
tween.TweenProperty(GetNode("Sprite"), "scale", Vector2.Zero, 1.0f).SetTrans(Tween.TransitionType.Bounce);
tween.TweenCallback(Callable.From(GetNode("Sprite").QueueFree));
[/csharp]
*/

package main

import (
	"github.com/AveryLucas/gogogd/classdb/Node"
	"github.com/AveryLucas/gogogd/classdb/PropertyTweener"
	"github.com/AveryLucas/gogogd/classdb/SceneTree"
	"github.com/AveryLucas/gogogd/classdb/Sprite2D"
	"github.com/AveryLucas/gogogd/classdb/Tween"
	"github.com/AveryLucas/gogogd/variant/Color"
	"github.com/AveryLucas/gogogd/variant/Vector2"
)

func ExampleTweenSetTrans(node Node.Instance, sprite Sprite2D.Instance) {
	var tween = SceneTree.Get(node).CreateTween()
	PropertyTweener.Make(tween, sprite.AsObject(), "modulate", Color.W3C.Red, 1).SetTrans(Tween.TransSine)
	PropertyTweener.Make(tween, sprite.AsObject(), "scale", Vector2.Zero, 1).SetTrans(Tween.TransBounce)
	tween.TweenCallback(sprite.AsNode().QueueFree)
}
