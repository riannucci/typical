package typical

import (
	"reflect"
)

type errorValue reflect.Value

var _ Value = errorValue{}

func (e errorValue) S(consumeFuncs ...interface{}) Value {
	ret := doSwitch([]reflect.Value{(reflect.Value)(e)}, consumeFuncs)
	if ret == nil {
		ret = e
	}
	return ret
}

func (e errorValue) First() interface{} {
	panic((reflect.Value)(e).Interface())
}

func (e errorValue) FirstErr() (interface{}, error) {
	return nil, e.Error()
}

func (e errorValue) All() []interface{} {
	panic((reflect.Value)(e).Interface())
}

func (e errorValue) Error() error {
	return (reflect.Value)(e).Interface().(error)
}
