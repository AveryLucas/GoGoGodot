/*
var texture = NoiseTexture3D.new()
texture.noise = FastNoiseLite.new()
await texture.changed
var data = texture.get_data()
*/

package main

import (
	"github.com/AveryLucas/gogogd/classdb/FastNoiseLite"
	"github.com/AveryLucas/gogogd/classdb/NoiseTexture3D"
)

func ExampleNoiseTexture3D() {
	var texture = NoiseTexture3D.New()
	texture.SetNoise(FastNoiseLite.New().AsNoise())
	texture.AsResource().OnChanged(func() {
		var data = texture.AsTexture3D().GetData()
		_ = data
	})
}
