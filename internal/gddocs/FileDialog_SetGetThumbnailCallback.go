/*
func _ready():
	FileDialog.set_get_thumbnail_callback(thumbnail_method)

func thumbnail_method(path):
	var image_texture = ImageTexture.new()
	make_thumbnail_async(path, image_texture)
	return image_texture

func make_thumbnail_async(path, image_texture):
	var thumbnail_texture = await generate_thumbnail(path) # Some method that generates a thumbnail.
	image_texture.set_image(thumbnail_texture.get_image())
*/

package main

import (
	"github.com/AveryLucas/gogogd/classdb/FileDialog"
	"github.com/AveryLucas/gogogd/classdb/ImageTexture"
	"github.com/AveryLucas/gogogd/classdb/Texture2D"
	"github.com/AveryLucas/gogogd/variant/Object"
)

type Thumbnails struct {
	Object.Extension[Thumbnails]
}

func (Thumbnails) Ready() {
	FileDialog.SetGetThumbnailCallback(thumbnail_method)
}

func thumbnail_method(path string) Texture2D.Instance {
	var image_texture = ImageTexture.New()
	return image_texture.AsTexture2D()
}
