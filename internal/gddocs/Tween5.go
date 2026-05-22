/*
[gdscript]
var tween
func animate():
	if tween:
		tween.kill() # Abort the previous animation.
	tween = create_tween()
[/gdscript]
[csharp]
private Tween _tween;

public void Animate()
{
	if (_tween != null)
		_tween.Kill(); // Abort the previous animation
	_tween = CreateTween();
}
[/csharp]
*/

package main

import (
	"github.com/AveryLucas/gogogd/classdb/Node"
	"github.com/AveryLucas/gogogd/classdb/Tween"
)

var tween Tween.Instance

func Animate(node Node.Instance) {
	if tween != Tween.Nil {
		tween.Kill()
	}
	tween = node.CreateTween()
}
