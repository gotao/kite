package dnode

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Callback is the callable function with arbitrary args and no return value.
type Callback func(args ...interface{})

// Path represents a callback function's path in the arguments structure.
// Contains mixture of string and integer values.
type Path []interface{}

// CallbackSpec is a structure encapsulating a Callback and it's Path.
type CallbackSpec struct {
	// Path represents the callback's path in the arguments structure.
	Path     Path
	Callback Callback
}

func (c *CallbackSpec) Apply(value reflect.Value) error {
	l.Printf("Apply: %#v\n", value.Interface())
	i := 0
	for {
		switch value.Kind() {
		case reflect.Slice:
			if i == len(c.Path) {
				return fmt.Errorf("Callback path too short: %v", c.Path)
			}

			// Path component may be a string or an integer.
			var index int
			var err error
			switch v := c.Path[i].(type) {
			case string:
				index, err = strconv.Atoi(v)
				if err != nil {
					return fmt.Errorf("Integer expected in callback path, got '%v'.", c.Path[i])
				}
			case int:
				index = v
			default:
				panic(fmt.Errorf("Unknown type: %#v", c.Path[i]))
			}

			value = value.Index(index)
			i++
		case reflect.Map:
			if i == len(c.Path) {
				return fmt.Errorf("Callback path too short: %v", c.Path)
			}
			if i == len(c.Path)-1 && value.Type().Elem().Kind() == reflect.Interface {
				value.SetMapIndex(reflect.ValueOf(c.Path[i]), reflect.ValueOf(c.Callback))
				return nil
			}
			value = value.MapIndex(reflect.ValueOf(c.Path[i]))
			i++
		case reflect.Ptr:
			value = value.Elem()
		case reflect.Interface:
			if i == len(c.Path) {
				value.Set(reflect.ValueOf(c.Callback))
				return nil
			}
			value = value.Elem()
		case reflect.Struct:
			if innerPartial, ok := value.Addr().Interface().(*Partial); ok {
				innerPartial.CallbackSpecs = append(innerPartial.CallbackSpecs, CallbackSpec{c.Path[i:], c.Callback})
				return nil
			}

			// Path component may be a string or an integer.
			name, ok := c.Path[i].(string)
			if !ok {
				return fmt.Errorf("Invalid path: %#v", c.Path[i])
			}

			value = value.FieldByName(strings.ToUpper(name[0:1]) + name[1:])
			i++
		case reflect.Func:
			value.Set(reflect.ValueOf(c.Callback))
			return nil
		case reflect.Invalid:
			// callback path does not exist, skip
			return nil
		default:
			return fmt.Errorf("Unhandled value of kind '%v' in callback path.", value.Kind())
		}
	}
	return nil
}