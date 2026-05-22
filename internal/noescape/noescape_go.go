//go:build !cgo && !wasm

package noescape

import "github.com/AveryLucas/gogogd/internal/gdextension"

func Call[T any](object gdextension.Object, method gdextension.MethodForClass, shape gdextension.Shape, args any) T {
	panic("not implemented")
}

func (method MethodForClass) Call(self gdextension.Object, args ...gdextension.Variant) (gdextension.Variant, error) {
	panic("not implemented")
}
