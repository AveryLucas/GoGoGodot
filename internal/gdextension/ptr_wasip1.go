//go:build wasip1

package gdextension

import (
	"github.com/AveryLucas/gogogd/variant/Color"
	"github.com/AveryLucas/gogogd/variant/Vector2"
	"github.com/AveryLucas/gogogd/variant/Vector3"
	"github.com/AveryLucas/gogogd/variant/Vector4"
)

type Pointer = uintptr
type Returns[T any] uintptr
type Accepts[T any] uintptr
type PackedArray[T byte | int32 | int64 | float32 | float64 | Color.RGBA | Vector2.XY | Vector3.XYZ | Vector4.XYZW | String] [2]uint64

const (
	SizeString      Shape = ShapeBytes8
	SizeObject      Shape = ShapeBytes8
	SizeArray       Shape = ShapeBytes8
	SizePackedArray Shape = ShapeBytes8x2
	SizeDictionary  Shape = ShapeBytes8
	SizeStringName  Shape = ShapeBytes8
	SizeNodePath    Shape = ShapeBytes8
	SizePointer     Shape = ShapeBytes8
)
