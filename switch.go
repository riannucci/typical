package typical

import (
	"fmt"
	"reflect"
)

func (v *Value) matchErr(fnT reflect.Type) bool {
	if fnT.NumIn() != 1 {
		return false
	}
	inT := fnT.In(0)
	if !inT.Implements(typeOfError) {
		return false
	}

	inID := reflect.ValueOf(inT).Pointer()
	key := matchKey{v.dataID, inID}

	match, ok, fromTypes := getMatchData(key)
	if ok {
		return match
	}

	return setMatchMap(key, fromTypes[0].AssignableTo(inT))
}

func (v *Value) match(fnT reflect.Type) bool {
	vt := reflect.Type(nil)
	numIn := fnT.NumIn()
	if fnT.IsVariadic() {
		if numIn-1 > len(v.dataErr) {
			return false
		}
		vt = fnT.In(numIn - 1).Elem()
		numIn--
	} else if len(v.dataErr) != numIn {
		return false
	}

	fnID := reflect.ValueOf(fnT).Pointer()
	key := matchKey{v.dataID, fnID}
	match, ok, fromTypes := getMatchData(key)
	if ok {
		return match
	}

	for i, t := range fromTypes {
		inT := reflect.Type(nil)
		if i < numIn {
			inT = fnT.In(i)
		} else {
			if vt == typeOfInterface {
				// optimize for ...interface{}
				break
			}
			inT = vt
		}
		if t == inT || (t == nil && inT == typeOfInterface) || (t != nil && t.AssignableTo(inT)) {
			continue
		}
		return setMatchMap(key, false)
	}

	return setMatchMap(key, true)
}

// S does a type-switch on the data in this value. Each consumeFunc must be
// a function. Switch will select and execute the first function whose inputs
// match the data in this Value. Only one consumeFunc per S will ever be
// called.
//
// If this Value is in an error state, functions will match against the
// singular error value. This can be used to distinguish between multiple
// error types. A function consuming the interface type `error` will match
// any error type. Functions intended to consume errors must have a single
// argument, and that single argument must either be `error` or a type which
// implements error.
//
// If the consuming function has the signature `func(...) (..., error)`, and
// returns a non-nil error, the returned Value will be in an error state. Note
// that the last returned value MUST be exactly of type `error` (not simply
// something that implements the `error` interface).
//
// If no function signature matches, S will return itself. This means that
// data and errors will continue to propagate down a switch chain until some
// function matches either the data or the error.
//
// Panics are not handled specially; if a consumeFunc panics, it will
// propagate without any intervention (i.e. it won't be converted to an
// error-state Value or anything like that).
//
// If any value in (first, rest...) is not a function, this will panic.
func (v *Value) S(first interface{}, rest ...interface{}) *Value {
	matchFn := v.match
	if v.dataID.isErr() {
		matchFn = v.matchErr
	}

	fnV := reflect.ValueOf(first)
	fnT := fnV.Type()
	if matchFn(fnT) {
		return v.call(&fnV, fnT)
	}

	for _, fn := range rest {
		fnV := reflect.ValueOf(fn)
		fnT := fnV.Type()
		if matchFn(fnT) {
			return v.call(&fnV, fnT)
		}
	}

	return v
}

func notNillableOrNotNil(v *reflect.Value) bool {
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Slice:
		return !v.IsNil()
	}
	return true
}

var (
	typeOfError     = reflect.TypeOf((*error)(nil)).Elem()
	typeOfInterface = reflect.TypeOf((*interface{})(nil)).Elem()

	empty               = interface{}(nil)
	valueOfNilInterface = reflect.ValueOf(&empty).Elem()
)

var commonFunctions = map[reflect.Type]func(interface{}, []reflect.Value) []reflect.Value{}

func RegisterCommonFunction(fn interface{}, impl func(interface{}, []reflect.Value) []reflect.Value) {
	fnT := reflect.TypeOf(fn)
	if fnT.Kind() != reflect.Func {
		panic(fmt.Errorf("typical.RegisterCommonFunction: %T must be a function", fn))
	}
	commonFunctions[fnT] = impl
}

func IfaceToValues(data ...interface{}) []reflect.Value {
	dataVals := []reflect.Value(nil)
	if len(data) > 0 {
		dataVals = make([]reflect.Value, len(data))
		for i, v := range data {
			d := reflect.ValueOf(v)
			if !d.IsValid() {
				dataVals[i] = valueOfNilInterface
			} else {
				dataVals[i] = d
			}
		}
	}
	return dataVals
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
	RegisterCommonFunction((func(a interface{}) error)(nil), func(fnI interface{}, in []reflect.Value) []reflect.Value {
		f := fnI.(func(interface{}) error)
		return IfaceToValues(f(in[0].Interface()))
	})
	RegisterCommonFunction((func(a, b interface{}) error)(nil), func(fnI interface{}, in []reflect.Value) []reflect.Value {
		f := fnI.(func(a, b interface{}) error)
		return IfaceToValues(f(in[0].Interface(), in[1].Interface()))
	})
}

func (v *Value) call(fnV *reflect.Value, fnT reflect.Type) *Value {
	fn := (func([]reflect.Value) []reflect.Value)(nil)
	if cmn, ok := commonFunctions[fnT]; ok {
		fn = func(in []reflect.Value) []reflect.Value {
			return cmn(fnV.Interface(), v.dataErr)
		}
	} else {
		fn = fnV.Call
	}
	data := fn(v.dataErr)

	if len(data) == 0 {
		return newData(fnT, nil)
	}

	lastIdx := fnT.NumOut() - 1
	if fnT.Out(lastIdx) == typeOfError {
		if !data[lastIdx].IsNil() {
			return newError(data[lastIdx])
		}
		data = data[:lastIdx]
	}
	return newData(fnT, data)
}
