// Package typical is a promise-style type-switching library for golang.
//
// It's probably very silly, but I thought it would be a fun project, and
// thought that it could make some really ugly error-handling patterns in go
// much nicer looking (but due to reflection, probably inordinately slow :)).
package typical

import (
	"reflect"
)

// Value represents a collection of data, or an error (never both). The Value
// may be switched by a variety of functions to produce a new Value, or the
// data/error may be retrieved from this Value.
//
// When a Value has data (i.e. is not an error), it holds 0 or more values.
type Value struct {
	// contains the multiple data or the singular error
	dataErr []reflect.Value
	dataID  typeID
}

// First will return the first datum of this Value
//
// This will or panic if this Value is in an error state.
func (v *Value) First() interface{} {
	r, err := v.FirstErr()
	if err != nil {
		panic(err)
	}
	return r
}

// FirstErr will return the first datum of this Value or the error.
func (v *Value) FirstErr() (interface{}, error) {
	if err := v.Error(); err != nil {
		return nil, err
	}
	first := v.dataErr[0]
	if notNillableOrNotNil(&first) {
		return first.Interface(), nil
	}
	return nil, nil
}

// All will return all the data in this Value
//
// This will panic if this Value is in an error state.
func (v *Value) All() []interface{} {
	r, err := v.AllErr()
	if err != nil {
		panic(err)
	}
	return r
}

// AllErr will return the first datum of this Value or the error.
func (v *Value) AllErr() ([]interface{}, error) {
	if err := v.Error(); err != nil {
		return nil, err
	}
	ret := make([]interface{}, len(v.dataErr))
	for i, v := range v.dataErr {
		if notNillableOrNotNil(&v) {
			ret[i] = v.Interface()
		}
	}
	return ret, nil
}

// Error returns the current error if this Value is in an error state, or nil
// otherwise.
func (v *Value) Error() error {
	if v.dataID.isErr() {
		return v.dataErr[0].Interface().(error)
	}
	return nil
}

func newData(fnT reflect.Type, data []reflect.Value) *Value {
	return &Value{data, dataToTypeID(false, fnT, data)}
}

func newError(err reflect.Value) *Value {
	data := []reflect.Value{err}
	return &Value{data, dataToTypeID(true, nil, data)}
}

// Do takes a niladic function which returns data and/or an error. It will
// invoke the function, and return a Value containing either the data or the
// error.
//
// This will panic if `fn` is the wrong type.
func Do(fn interface{}) *Value {
	fnV := reflect.ValueOf(fn)
	return retDataToValue(fnV.Type(), fnV.Call(nil))
}

// Data creates a data-Value containing the provided data.
func Data(data ...interface{}) *Value {
	dataVals := make([]reflect.Value, len(data))
	for i, v := range data {
		d := reflect.ValueOf(v)
		if !d.IsValid() {
			dataVals[i] = valueOfNilInterface
		} else {
			dataVals[i] = d
		}
	}
	return newData(nil, dataVals)
}

// Error creates an error-Value containing the provided error.
//
// If the error is nil, this is equivalent to Data() (i.e. a niladic data-Value).
func Error(err error) *Value {
	if err == nil {
		return Data()
	}
	return newError(reflect.ValueOf(err))
}
