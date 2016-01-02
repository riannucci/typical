package typical

import (
	"fmt"
	"reflect"
)

var commonFunctions = map[reflect.Type]func(interface{}, []reflect.Value) []reflect.Value{}

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
	RegisterCommonFunction((func() int)(nil), func(fnI interface{}, in []reflect.Value) []reflect.Value {
		f := fnI.(func() int)
		return IfaceToValues(f())
	})
	RegisterCommonFunction((func(int))(nil), func(fnI interface{}, in []reflect.Value) []reflect.Value {
		f := fnI.(func(int))
		f(int(in[0].Int()))
		return nil
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
	RegisterCommonFunction((func(interface{}))(nil), func(fnI interface{}, in []reflect.Value) []reflect.Value {
		f := fnI.(func(interface{}))
		f(in[0].Interface())
		return nil
	})
	RegisterCommonFunction((func(a, b interface{}) error)(nil), func(fnI interface{}, in []reflect.Value) []reflect.Value {
		f := fnI.(func(a, b interface{}) error)
		return IfaceToValues(f(in[0].Interface(), in[1].Interface()))
	})
}
