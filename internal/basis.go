//go:build !generate

package gd

import (
	basis "github.com/AveryLucas/gogogd/variant/Basis"
)

func Transposed(t Transform3D) Transform3D {
	return Transform3D{
		Basis:  basis.Transposed(t.Basis),
		Origin: t.Origin,
	}
}
