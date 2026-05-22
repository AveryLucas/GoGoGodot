/*
var texture = load("res://icon.svg")
$Sprite2D.texture = texture
*/

package main

import (
	"github.com/AveryLucas/gogogd/classdb/Resource"
	"github.com/AveryLucas/gogogd/classdb/Sprite2D"
	"github.com/AveryLucas/gogogd/classdb/Texture2D"
)

func ExampleLoadImageTexture(sprite Sprite2D.Instance) {
	var texture = Resource.Load[Texture2D.Instance]("res://icon.svg")
	sprite.SetTexture(texture.AsTexture2D())
}
