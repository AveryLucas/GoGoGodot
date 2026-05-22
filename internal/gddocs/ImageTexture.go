/*
var image = Image.load_from_file("res://icon.svg")
var texture = ImageTexture.create_from_image(image)
$Sprite2D.texture = texture
*/

package main

import (
	"github.com/AveryLucas/gogogd/classdb/Image"
	"github.com/AveryLucas/gogogd/classdb/ImageTexture"
	"github.com/AveryLucas/gogogd/classdb/Sprite2D"
)

func ExampleImageTexture(sprite Sprite2D.Instance) {
	var image = Image.LoadFromFile("res://icon.svg")
	var texture = ImageTexture.CreateFromImage(image)
	sprite.SetTexture(texture.AsTexture2D())
}
