package gd_test

import (
	"testing"

	"github.com/AveryLucas/gogogd/classdb/GDScript"
	"github.com/AveryLucas/gogogd/variant/Angle"
	"github.com/AveryLucas/gogogd/variant/Basis"
	"github.com/AveryLucas/gogogd/variant/Euler"
	"github.com/AveryLucas/gogogd/variant/Object"

	_ "embed"
)

var basis_test string = `extends Object

func test_basis(angles):
    return Basis.from_euler(angles)
`

func TestBasis(t *testing.T) {
	runOnMain(t, func(t testing.TB) {
		var runner = Object.New()
		var script = GDScript.New().AsScript()
		script.SetSourceCode(basis_test)
		script.Reload()
		runner.SetScript(script)

		angles := Euler.Radians{X: -1, Y: 0.2, Z: 0}
		engine := Object.Call(runner, "test_basis", angles).(Basis.XYZ)
		if engine != Basis.FromEuler(angles, Angle.OrderYXZ) {
			t.Fatalf("Expected %v, got %v", Object.Call(runner, "test_basis", angles), Basis.FromEuler(angles, Angle.OrderYXZ))
		}
	})
}
