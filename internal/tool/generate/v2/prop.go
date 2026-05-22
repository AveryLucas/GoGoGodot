package main

import (
	"fmt"
	"io"

	"graphics.gd/internal/gdjson"
	"graphics.gd/internal/tool/generate/gdtype"
)

func (classDB ClassDB) new(file io.Writer, class gdjson.Class) {
	fmt.Fprintf(file, "func New() Instance {\n")
	fmt.Fprintf(file, `if !gd.Linked {
		var placeholder = Instance([1]gdclass.%[1]s{gdclass.New%[1]s(gdreference.NewObject())})
		gd.StartupFunctions = append(gd.StartupFunctions, func() {
			if gd.Linked {
				raw, _ := gdreference.EndObject(New().AsObject()[0])
				gdreference.SetObject(gdclass.Get%[1]s(placeholder[0])[0], raw)
				gd.RegisterCleanup(func() {
					if raw := gdreference.GetObject(placeholder.AsObject()[0]); raw != 0 {
						gdextension.Host.Objects.Unsafe.Free(raw)
					}
				})
			}
		})
		return placeholder
	}
`, class.Name)
	fmt.Fprintf(file, "\tcasted := Instance([1]gdclass.%[1]s{gdclass.New%[1]s(gdreference.OwnObject(gdextension.Host.Objects.Make(sname), gd.Free))})\n", class.Name)
	if class.IsRefcounted {
		fmt.Fprintf(file, "\tcasted.AsRefCounted()[0].InitRef()\n")
	}
	fmt.Fprintf(file, "\tgd.ObjectNotification(casted.AsObject()[0], 0, false)\n")
	fmt.Fprintf(file, "\treturn casted\n")
	fmt.Fprintf(file, "}\n")
}

// propRef is the minimal subset of a class property we use during
// promotion. The upstream Property type is an anonymous struct inside
// gdjson.Class — we can't import it by name, so we extract just what
// we need.
type propRef struct {
	Name   string
	Type   string
	Setter string
	Getter string
	Index  *int
}

// promotedPropertyAccessor emits getter (and setter where applicable)
// forwarders on the leaf's *Extension[T] for a property defined on an
// ancestor class. Skips emission if the leaf already has a same-named
// method/property on *Extension[T].
//
// Each forwarder delegates through the AsAncestor() chain. The setter
// forwarder returns *Extension[T] (the leaf's) for chaining consistency.
func (classDB ClassDB) promotedPropertyAccessor(file io.Writer, leafClass gdjson.Class, ancestor gdjson.Class, prop propRef, alreadyEmitted map[string]bool) {
	getterName := convertName(prop.Name)
	setterName := "Set" + getterName

	// Resolve the property type the same way the original `properties`
	// function does: start with a lookup using prop.Name, then refine
	// using the getter or setter method's return/arg type when available.
	// This ensures Distinctions-driven renames (e.g. int → RenderPriority
	// for Material.render_priority) match correctly.
	// Skip if the getter or setter is relocated to another class — those
	// don't get emitted on the source Instance, so the leaf can't forward
	// to them. Mirrors the same skip in `properties`.
	if prop.Getter != "" {
		if _, reloc := gdjson.Relocations[ancestor.Name+"."+prop.Getter]; reloc {
			return
		}
	}
	if prop.Setter != "" {
		if _, reloc := gdjson.Relocations[ancestor.Name+"."+prop.Setter]; reloc {
			return
		}
	}

	ptype := classDB.convertTypeSimple(ancestor, ancestor.Name+"."+prop.Name, "", prop.Type)
	var hasGetter, hasSetter bool
	if prop.Getter != "" {
		for _, m := range ancestor.Methods {
			if m.Name == prop.Getter {
				ptype = classDB.convertTypeSimple(ancestor, ancestor.Name+"."+prop.Getter+".", m.ReturnValue.Meta, m.ReturnValue.Type)
				hasGetter = true
				break
			}
		}
	}
	if prop.Setter != "" {
		for _, m := range ancestor.Methods {
			if m.Name == prop.Setter {
				idx := 0
				if prop.Index != nil {
					idx = 1
				}
				if idx >= len(m.Arguments) {
					break
				}
				ptype = classDB.convertTypeSimple(ancestor, ancestor.Name+"."+prop.Setter+"."+prop.Name, m.Arguments[idx].Meta, m.Arguments[idx].Type)
				hasSetter = true
				break
			}
		}
	}
	ptype = qualifyForLeaf(ptype, ancestor.Name, leafClass.Name)
	if !hasGetter && !hasSetter {
		return
	}

	if hasGetter {
		// Promote onto Instance (delegates via AsAncestor()).
		key := getterName + "::Instance"
		if !alreadyEmitted[key] {
			alreadyEmitted[key] = true
			fmt.Fprintf(file, "\n// %s is promoted from [%s.Instance.%s].\n", getterName, ancestor.Name, getterName)
			fmt.Fprintf(file, "func (self Instance) %s() %s { return self.As%s().%s() }\n",
				getterName, ptype, ancestor.Name, getterName)
		}
		// Promote onto *Extension[T] (delegates via Super().AsAncestor()).
		key = getterName + "::*Extension[T]"
		if !alreadyEmitted[key] {
			alreadyEmitted[key] = true
			fmt.Fprintf(file, "func (o *Extension[T]) %s() %s { return o.Super().As%s().%s() }\n",
				getterName, ptype, ancestor.Name, getterName)
		}
	}
	if hasSetter {
		key := setterName + "::Instance"
		if !alreadyEmitted[key] {
			alreadyEmitted[key] = true
			fmt.Fprintf(file, "\n// %s is promoted from [%s.Instance.%s].\n", setterName, ancestor.Name, setterName)
			fmt.Fprintf(file, "func (self Instance) %s(value %s) Instance {\n", setterName, ptype)
			fmt.Fprintf(file, "\tself.As%s().%s(value)\n", ancestor.Name, setterName)
			fmt.Fprintf(file, "\treturn self\n}\n")
		}
		key = setterName + "::*Extension[T]"
		if !alreadyEmitted[key] {
			alreadyEmitted[key] = true
			fmt.Fprintf(file, "func (o *Extension[T]) %s(value %s) *Extension[T] {\n", setterName, ptype)
			fmt.Fprintf(file, "\to.Super().As%s().%s(value)\n", ancestor.Name, setterName)
			fmt.Fprintf(file, "\treturn o\n}\n")
		}
	}
}

func (classDB ClassDB) properties(file io.Writer, class gdjson.Class, singleton bool) {
	if len(class.Properties) == 0 {
		return
	}
	for _, prop := range class.Properties {
		ptype := classDB.convertTypeSimple(class, class.Name+"."+prop.Name, "", prop.Type)
		expert := gdtype.EngineTypeAsGoType(class.Name, "", prop.Type)
		var foundGetter bool
		var foundSetter bool
		if prop.Getter != "" {
			for _, method := range class.Methods {
				if gdjson.Relocations[class.Name+"."+method.Name] != "" {
					continue
				}
				if method.Name == prop.Getter {
					ptype = classDB.convertTypeSimple(class, class.Name+"."+prop.Getter+".", method.ReturnValue.Meta, method.ReturnValue.Type)
					foundGetter = true
					expert = gdtype.EngineTypeAsGoType(class.Name, method.ReturnValue.Meta, method.ReturnValue.Type)
					break
				}
			}
		}
		if prop.Setter != "" {
			for _, method := range class.Methods {
				if gdjson.Relocations[class.Name+"."+method.Name] != "" {
					continue
				}
				if method.Name == prop.Setter {
					var i = 0
					if prop.Index != nil {
						i = 1
					}
					ptype = classDB.convertTypeSimple(class, class.Name+"."+prop.Setter+"."+prop.Name, method.Arguments[i].Meta, method.Arguments[i].Type)
					expert = gdtype.EngineTypeAsGoType(class.Name, method.Arguments[i].Meta, method.Arguments[i].Type)
					foundSetter = true
					break
				}
			}
		}
		if !foundGetter && !foundSetter {
			continue
		}
		if foundGetter {
			if prop.Description != "" {
				fmt.Fprintln(file, "\n/*")
				fmt.Fprint(file, gdjson.DocsToGoDoc(prop.Description, classDB, class.Name, class.Name+"_"+convertName(prop.Name)))
				fmt.Fprint(file, "\n*/")
			}
			if singleton {
				fmt.Fprintf(file, "\nfunc %s() %s { //gd:%s.%s\n", convertName(prop.Name), ptype, class.Name, prop.Name)
				fmt.Fprintf(file, "once.Do(singleton)\n\t")
			} else {
				fmt.Fprintf(file, "\nfunc (self Instance) %s() %s { //gd:%s.%s\n", convertName(prop.Name), ptype, class.Name, prop.Name)
			}
			val := fmt.Sprintf("class(self).%s()", convertName(prop.Getter))
			if prop.Index != nil {
				val = fmt.Sprintf("class(self).%s(%d)", convertName(prop.Getter), *prop.Index)
			}
			fmt.Fprintf(file, "\t\treturn %s(%s)\n", ptype, gdtype.Name(expert).ConvertToGo(val, ptype))
			fmt.Fprintf(file, "}\n")
			// gogogd-fork: also expose the property getter on *Extension[T]
			// so users who embed Foo.Extension[T] can call p.Foo() directly
			// without going through p.Super().Foo().
			if !singleton {
				fmt.Fprintf(file, "\nfunc (o *Extension[T]) %s() %s { return o.Super().%s() }\n",
					convertName(prop.Name), ptype, convertName(prop.Name))
			}
		}

		if prop.Setter != "" {
			var found = true
			for _, method := range class.Methods {
				if convertName(method.Name) == convertName(prop.Setter) && method.Name != prop.Setter {
					found = false
					break
				}
			}
			if !found {
				continue
			}
			found = false
			for _, method := range class.Methods {
				if method.Name == prop.Setter {
					var i = 0
					if prop.Index != nil {
						i = 1
					}
					ptype = classDB.convertTypeSimple(class, class.Name+"."+prop.Setter+"."+prop.Name, method.Arguments[i].Meta, method.Arguments[i].Type)
					expert = gdtype.EngineTypeAsGoType(class.Name, method.Arguments[i].Meta, method.Arguments[i].Type)
					found = true
					break
				}
			}
			if !found {
				continue
			}
			if !foundGetter {
				if prop.Description != "" {
					fmt.Fprintln(file, "\n/*")
					fmt.Fprint(file, gdjson.DocsToGoDoc(prop.Description, classDB, class.Name, class.Name+"_"+convertName(prop.Name)))
					if !singleton {
						fmt.Fprintf(file, "\nReturns the instance, so that property settings can be chained.")
					}
					fmt.Fprint(file, "\n*/")
				}
			} else {
				fmt.Fprintf(file, "\n// Set%s sets the property returned by [%s].", convertName(prop.Name), convertName(prop.Getter))
				if !singleton {
					fmt.Fprintf(file, " Returns the instance, so that property settings can be chained.")
				}
			}
			if singleton {
				fmt.Fprintf(file, "\nfunc Set%s(value %s) { //gd:%s.%s\n", convertName(prop.Name), ptype, class.Name, prop.Name)
				fmt.Fprintf(file, "once.Do(singleton)\n\t")
			} else {
				fmt.Fprintf(file, "\nfunc (self Instance) Set%s(value %s) Instance { //gd:%s.%s\n", convertName(prop.Name), ptype, class.Name, prop.Name)
			}
			if prop.Index != nil {
				fmt.Fprintf(file, "\tclass(self).%s(%d, %s)\n", convertName(prop.Setter), *prop.Index, gdtype.Name(expert).ConvertToSimple("value", ptype))
			} else {
				fmt.Fprintf(file, "\tclass(self).%s(%s)\n", convertName(prop.Setter), gdtype.Name(expert).ConvertToSimple("value", ptype))
			}
			if !singleton {
				fmt.Fprintf(file, "\treturn self\n")
			}
			fmt.Fprintf(file, "}\n")
			// gogogd-fork: also expose the property setter on
			// *Extension[T]. Returns *Extension[T] for chaining (the
			// leaf's, not the parent's).
			if !singleton {
				fmt.Fprintf(file, "\nfunc (o *Extension[T]) Set%s(value %s) *Extension[T] {\n", convertName(prop.Name), ptype)
				fmt.Fprintf(file, "\to.Super().Set%s(value)\n", convertName(prop.Name))
				fmt.Fprintf(file, "\treturn o\n}\n")
			}
		}
	}
}
