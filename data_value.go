package typical

import (
	"reflect"
)

type dataValue []reflect.Value

var _ Value = dataValue(nil)

func (d dataValue) S(consumeFuncs ...interface{}) Value {
	ret := doSwitch(d, consumeFuncs)
	if ret == nil {
		ret = d
	}
	return ret
}

func (d dataValue) First() interface{} {
	if !d[0].IsNil() {
		return d[0].Interface()
	}
	return nil
}

func (d dataValue) All() []interface{} {
	ret := make([]interface{}, len(d))
	for i, v := range d {
		if !v.IsNil() {
			ret[i] = v.Interface()
		}
	}
	return ret
}

func (d dataValue) Error() error {
	return nil
}
