/*
var tween = get_tree().create_tween()
tween.tween_callback(queue_free).set_delay(2)
*/

package main

import (
	"github.com/AveryLucas/gogogd/classdb/Node"
	"github.com/AveryLucas/gogogd/classdb/SceneTree"
)

var node Node.Instance

func CallbackTweener_SetDelay() {
	var tween = SceneTree.Get(node).CreateTween()
	tween.TweenCallback(node.QueueFree).SetDelay(2)
}
