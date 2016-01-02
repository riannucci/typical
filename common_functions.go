package typical

import (
	"fmt"
	"reflect"
)

var commonFunctions = map[reflect.Type]func(interface{}, []reflect.Value) []reflect.Value{}

// RegisterCommonFunction allows you to optimize typical's function call
// mechanism. Without registering a function, typical uses reflect.Call to
// invoke functions. However, if you register a function signature, typical
// can invoke this function type more directly, which is much faster, and
// cuts down on a lot of garbage collection junk.
//
// Example:
//   RegisterCommonFunction((func(string) error)(nil), func(fn interface{}, args []reflect.Value) (out []reflect.Value) {
//     return IfaceToValues(fn.(func(string) error)(args[0].Interface().(string)))
//   })
//
// The following function signatures are implemented by default:
//   func(interface{}) ([]byte, error)
//   func([]byte) (int, error)
//   func() (error)
//   func(error)
//   func()
//   func(...interface{}) error
func RegisterCommonFunction(fn interface{}, impl func(interface{}, []reflect.Value) []reflect.Value) {
	fnT := reflect.TypeOf(fn)
	if fnT.Kind() != reflect.Func {
		panic(fmt.Errorf("typical.RegisterCommonFunction: %T must be a function", fn))
	}
	commonFunctions[fnT] = impl
}

func init() {
	RegisterCommonFunction((func(interface{}) ([]byte, error))(nil), func(fnI interface{}, in []reflect.Value) []reflect.Value {
		f := fnI.(func(interface{}) ([]byte, error))
		return IfaceToValues(f(in[0].Interface()))
	})
	RegisterCommonFunction((func([]byte) (int, error))(nil), func(fnI interface{}, in []reflect.Value) []reflect.Value {
		f := fnI.(func([]byte) (int, error))
		return IfaceToValues(f(in[0].Bytes()))
	})
	RegisterCommonFunction((func() error)(nil), func(fnI interface{}, in []reflect.Value) []reflect.Value {
		f := fnI.(func() error)
		return IfaceToValues(f())
	})
	RegisterCommonFunction((func(error))(nil), func(fnI interface{}, in []reflect.Value) []reflect.Value {
		f := fnI.(func(error))
		f(in[0].Interface().(error))
		return nil
	})
	RegisterCommonFunction((func())(nil), func(fnI interface{}, in []reflect.Value) []reflect.Value {
		f := fnI.(func())
		f()
		return nil
	})
	RegisterCommonFunction((func(...interface{}) error)(nil), func(fnI interface{}, in []reflect.Value) []reflect.Value {
		f := fnI.(func(...interface{}) error)
		inVals := make([]interface{}, len(in))
		for i, v := range in {
			inVals[i] = v.Interface()
		}
		return IfaceToValues(f(inVals...))
	})
}
