//go:build !generate

package gd

import (
	"iter"
	"unsafe"

	"github.com/AveryLucas/gogogd/internal/gdextension"
	"github.com/AveryLucas/gogogd/internal/gdreference"
	"github.com/AveryLucas/gogogd/internal/noescape"
	"github.com/AveryLucas/gogogd/internal/pointers"
	"github.com/AveryLucas/gogogd/internal/threadcheck"
	VariantPkg "github.com/AveryLucas/gogogd/variant"
	CallableType "github.com/AveryLucas/gogogd/variant/Callable"
	DictionaryType "github.com/AveryLucas/gogogd/variant/Dictionary"
	ErrorType "github.com/AveryLucas/gogogd/variant/Error"
	SignalType "github.com/AveryLucas/gogogd/variant/Signal"
	StringType "github.com/AveryLucas/gogogd/variant/String"
)

func (s Signal) Free() {
	if ptr, ok := pointers.End(s); ok {
		noescape.Free(gdextension.TypeSignal, &ptr)
	}
}

func NewSignalOf(object [1]gdreference.Object, signal StringName) Signal {
	return pointers.New[Signal](noescape.Make[gdextension.Signal](builtin.creation.Signal[2], gdextension.SizeObject<<4|gdextension.SizeStringName<<8, unsafe.Pointer(&struct {
		object gdextension.Object
		signal gdextension.StringName
	}{
		object: gdreference.GetObject(gdreference.Object(object[0])),
		signal: pointers.Get(signal),
	})))
}

func InternalSignal(signal SignalType.Any) Signal {
	_, state := SignalType.Proxy(signal, NewSignalCheck, NewSignalProxy)
	return pointers.Load[Signal](state)
}

type SignalProxy struct{}

func NewSignalProxy() (SignalProxy, complex128) {
	panic("NewSignalProxy: not implemented")
}

func NewSignalCheck(SignalProxy, complex128) bool {
	return true
}

func (SignalProxy) Attach(raw complex128, fn CallableType.Function, flags SignalType.Flags) error {
	sig := pointers.Load[Signal](raw)
	return ToError(ErrorType.Code(sig.Connect(InternalCallable(fn), Int(flags))))
}
func (SignalProxy) Remove(raw complex128, fn CallableType.Function) {
	sig := pointers.Load[Signal](raw)
	sig.Disconnect(InternalCallable(fn))
}
func (SignalProxy) Name(raw complex128) StringType.Unicode {
	sig := pointers.Load[Signal](raw)
	return StringType.New(sig.GetName().String())
}
func (SignalProxy) Consumers(raw complex128) iter.Seq[SignalType.Consumer] {
	sig := pointers.Load[Signal](raw)
	return func(fn func(SignalType.Consumer) bool) {
		for _, conn := range sig.GetConnections().Iter() {
			if !fn(DictionaryAs[SignalType.Consumer](conn.Interface().(DictionaryType.Any))) {
				break
			}
		}
	}
}
func (SignalProxy) Emit(raw complex128, values ...VariantPkg.Any) {
	if threadcheck.Main() {
		sig := pointers.Load[Signal](raw)
		vargs := make([]Variant, len(values))
		for i, v := range values {
			vargs[i] = NewVariant(v)
		}
		sig.Emit(vargs...)
		return
	}
	CallableType.Defer(CallableType.New(func() {
		sig := pointers.Load[Signal](raw)
		vargs := make([]Variant, len(values))
		for i, v := range values {
			vargs[i] = NewVariant(v)
		}
		sig.Emit(vargs...)
	}))
}
func (SignalProxy) Emitter(raw complex128) VariantPkg.Any {
	return VariantPkg.New(gdreference.GetObject(gdreference.Object(pointers.Load[Signal](raw).GetObject())))
}
