package typical

import (
	"reflect"
)

var (
	empty               = interface{}(nil)
	valueOfNilInterface = reflect.ValueOf(&empty).Elem()
)

// IfaceToValues is a helper function to convert various values to a slice
// of reflect.Value.
func IfaceToValues(data ...interface{}) []reflect.Value {
	dataVals := []reflect.Value(nil)
	if len(data) > 0 {
		dataVals = make([]reflect.Value, len(data))
		for i, v := range data {
			dataVals[i] = reflect.ValueOf(v)
			if !dataVals[i].IsValid() {
				dataVals[i] = valueOfNilInterface
			}
		}
	}
	return dataVals
}

// Value represents a collection of data, or an error (never both). The Value
// may be switched by a variety of functions to produce a new Value, or the
// data/error may be retrieved from this Value.
//
// When a Value has data (i.e. is not an error), it holds 0 or more values.
type Value struct {
	// contains the multiple data or the singular error
	dataErr []reflect.Value
	isErr   bool
	//dataID  typeID
}

// First will return the first datum of this Value
//
// This will or panic if this Value is in an error state.
func (v Value) First() interface{} {
	r, err := v.FirstErr()
	if err != nil {
		panic(err)
	}
	return r
}

func notNillableOrNotNil(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Slice:
		return !v.IsNil()
	}
	return true
}

// FirstErr will return the first datum of this Value or the error.
func (v Value) FirstErr() (interface{}, error) {
	if err := v.Error(); err != nil {
		return nil, err
	}
	first := v.dataErr[0]
	if notNillableOrNotNil(first) {
		return first.Interface(), nil
	}
	return nil, nil
}

// All will return all the data in this Value
//
// This will panic if this Value is in an error state.
func (v Value) All() []interface{} {
	r, err := v.AllErr()
	if err != nil {
		panic(err)
	}
	return r
}

// AllErr will return the first datum of this Value or the error.
func (v Value) AllErr() ([]interface{}, error) {
	if err := v.Error(); err != nil {
		return nil, err
	}
	ret := make([]interface{}, len(v.dataErr))
	for i, v := range v.dataErr {
		if notNillableOrNotNil(v) {
			ret[i] = v.Interface()
		}
	}
	return ret, nil
}

// Error returns the current error if this Value is in an error state, or nil
// otherwise.
func (v Value) Error() error {
	if v.isErr {
		return v.dataErr[0].Interface().(error)
	}
	return nil
}

func newData(data []reflect.Value) Value {
	return Value{data, false}
}

func newError(err reflect.Value) Value {
	return Value{[]reflect.Value{err}, true}
}

// Do takes a niladic function which returns data and/or an error. It will
// invoke the function, and return a Value containing either the data or the
// error.
//
// This will panic if `fn` is the wrong type.
func Do(fn interface{}) Value {
	fnV := reflect.ValueOf(fn)
	return newData(nil).call(fnV, fnV.Type())
}

// Data creates a data-Value containing the provided data.
func Data(data ...interface{}) Value {
	return newData(IfaceToValues(data...))
}

// Error creates an error-Value containing the provided error.
//
// If the error is nil, this is equivalent to Data() (i.e. a niladic data-Value).
func Error(err error) Value {
	if err == nil {
		return Data()
	}
	return newError(reflect.ValueOf(err))
}
