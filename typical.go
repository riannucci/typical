// Package typical is a promise-style type-switching library for golang.
//
// It's probably very silly, but I thought it would be a fun project, and
// thought that it could make some really ugly error-handling patterns in go
// much nicer looking (but due to reflection, probably inordinately slow :)).
package typical

import (
	"fmt"
	"reflect"
)

// Value represents a collection of data, or an error (never both). The Value
// may be switched by a variety of functions to produce a new Value, or the
// data/error may be retrieved from this Value.
//
// When a Value has data (i.e. is not an error), it holds 0 or more values.
type Value interface {
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
	// If any value in consumeFuncs is not a function, this will panic.
	S(consumeFuncs ...interface{}) Value

	// FirstErr will return the first datum of this Value or the error.
	FirstErr() (interface{}, error)

	// FirstErr will return the first datum of this Value or the error.
	AllErr() ([]interface{}, error)

	// First will return the first datum of this Value
	//
	// This will or panic if this Value is in an error state.
	First() interface{}

	// All will return all the data in this Value
	//
	// This will panic if this Value is in an error state.
	All() []interface{}

	// Error returns the current error if this Value is in an error state, or nil
	// otherwise.
	Error() error
}

// Do takes a niladic function which returns data and/or an error. It will
// invoke the function, and return a Value containing either the data or the
// error.
//
// This will panic if `fn` is the wrong type.
func Do(fn interface{}) Value {
	fnV, fnT := callable(fn)
	if fnT.NumIn() != 0 {
		panic(fmt.Errorf("typical.Do: %T is not niladic", fn))
	}

	return retDataToValue(fnT, fnV.Call(nil))
}

// Data creates a data-Value containing the provided data.
func Data(data ...interface{}) Value {
	dataVals := make([]reflect.Value, len(data))
	for i, v := range data {
		dataVals[i] = reflect.ValueOf(v)
	}
	return newData(dataVals)
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
