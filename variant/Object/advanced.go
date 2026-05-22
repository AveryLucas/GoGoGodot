package Object

import (
	"reflect"
	"unsafe"

	"github.com/AveryLucas/gogogd/internal/gdclass"
	"github.com/AveryLucas/gogogd/internal/gdreference"
)

type Advanced [1]gdclass.Object

func (obj Advanced) AsObject() [1]gdreference.Object { return obj }
func (self *Advanced) UnsafePointer() unsafe.Pointer { return unsafe.Pointer(self) }

// Virtual method lookup.
func (obj Advanced) Virtual(name string) reflect.Value { return reflect.Value{} }
