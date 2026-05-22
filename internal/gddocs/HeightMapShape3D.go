/*
[gdscript]
var heightmap_texture = ResourceLoader.load("res://heightmap_image.exr")
var heightmap_image = heightmap_texture.get_image()
heightmap_image.convert(Image.FORMAT_RF)

var height_min = 0.0
var height_max = 10.0

update_map_data_from_image(heightmap_image, height_min, height_max)
[/gdscript]
*/

package main

import (
	"github.com/AveryLucas/gogogd/classdb/HeightMapShape3D"
	"github.com/AveryLucas/gogogd/classdb/Image"
	"github.com/AveryLucas/gogogd/classdb/Resource"
	"github.com/AveryLucas/gogogd/classdb/Texture2D"
	"github.com/AveryLucas/gogogd/variant/Float"
)

func ExampleHeightMapShape3D(shape HeightMapShape3D.Instance) {
	var heightmap_texture = Resource.Load[Texture2D.Instance]("res://heightmap_image.exr")
	var heightmap_image = heightmap_texture.GetImage()
	heightmap_image.Convert(Image.FormatRf)
	var height_min Float.X = 0.0
	var height_max Float.X = 10.0
	shape.UpdateMapDataFromImage(heightmap_image, height_min, height_max)
}
