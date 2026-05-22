/*
[gdscript]
var tween = get_tree().create_tween().bind_node(self).set_trans(Tween.TRANS_ELASTIC)
tween.tween_property($Sprite, "modulate", Color.RED, 1)
tween.tween_property($Sprite, "scale", Vector2(), 1)
tween.tween_callback($Sprite.queue_free)
[/gdscript]
[csharp]
var tween = GetTree().CreateTween().BindNode(this).SetTrans(Tween.TransitionType.Elastic);
tween.TweenProperty(GetNode("Sprite"), "modulate", Colors.Red, 1.0f);
tween.TweenProperty(GetNode("Sprite"), "scale", Vector2.Zero, 1.0f);
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

func ExampleTweenBound(node Node.Instance, sprite Sprite2D.Instance) {
	var tween = Tween.Instance(Tween.Advanced(SceneTree.Get(node).CreateTween()).BindNode(node)).SetTrans(Tween.TransElastic)
	PropertyTweener.Make(tween, sprite.AsObject(), "modulate", Color.W3C.Red, 1)
	PropertyTweener.Make(tween, sprite.AsObject(), "scale", Vector2.Zero, 1)
	tween.TweenCallback(sprite.AsNode().QueueFree)
}
