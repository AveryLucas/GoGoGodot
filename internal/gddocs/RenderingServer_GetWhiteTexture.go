/*
var texture_rid = RenderingServer.get_white_texture()
var texture = ImageTexture.create_from_image(RenderingServer.texture_2d_get(texture_rid))
$Sprite2D.texture = texture
*/

package main

import (
	"github.com/AveryLucas/gogogd/classdb/ImageTexture"
	"github.com/AveryLucas/gogogd/classdb/RenderingServer"
	"github.com/AveryLucas/gogogd/variant/RID"
)

func RenderingServer_GetWhiteTexture() {
	var texture_rid = RenderingServer.GetWhiteTexture()
	var texture = ImageTexture.CreateFromImage(RenderingServer.Texture2dGet(RID.Texture2D(texture_rid)))
	sprite.SetTexture(texture.AsTexture2D())
}
