package Script

import (
	gd "github.com/AveryLucas/gogogd/internal"
	"github.com/AveryLucas/gogogd/variant/Object"
)

// Get returns the object's Script instance, or false if no script is attached.
func Get(obj Object.Any) (Instance, bool) { //gd:Object.get_script
	script, ok := gd.ObjectGetScript(obj.AsObject()[0]).Interface().(Instance)
	return script, ok
}
