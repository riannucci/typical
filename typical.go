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
// may be Switched by a variety of functions to produce a new Value, or the
// data/error may be retrieved from this Value
type Value interface {
	// S does a type-switch on the data in this value. Each consumeFunc must be
	// a function. Switch will select and execute the first function whose inputs
	// match the data in this Value. Only one consumeFunc per S will ever be
	// called.
	//
	// If this Value is in an error state, functions will match against the
	// singular error value. This can be used to distinguish between multiple
	// error types. A function consuming the interface type `error` will match
	// any error type.
	//
	// If the consuming function has the signature `func(...) (..., error)`, and
	// returns a non-nil error, the returned Value will be in an error state. Note
	// that the last returned value MUST be exactly of type `error` (not simply
	// something that implements the `error` interface).
	//
	// If no function signature matches, S will simply return itself. This means
	// that data and errors will continue to propagate down a S chain until some
	// function matches either the data or the error.
	//
	// You can match a nil data value by having a consumeFunc with an argument
	// of NilValue. Otherwise a typeless nil will only be matchable by a function
	// taking `interface{}`.
	//
	// Panics are not handled specially; if a consumeFunction panics, typical
	// will propagate it without any intervention (i.e. it won't be converted
	// to an error-state Value or anything like that).
	S(consumeFuncs ...interface{}) Value

	// First will return the first datum of this Value, or panic if this Value is
	// in an error state.
	First() interface{}

	// FirstErr will return the first datum of this Value or the error.
	FirstErr() (interface{}, error)

	// All will return all the data in this Value, or panic if this Value is in an
	// error state.
	All() []interface{}

	// Error returns the current error if this Value is in an error state, or nil
	// otherwise.
	Error() error
}

type NilValue struct{}

var (
	typeOfError     = reflect.TypeOf((*error)(nil)).Elem()
	typeOfNilValue  = reflect.TypeOf(NilValue{})
	valueOfNilValue = reflect.ValueOf(NilValue{})
)

// Do takes a niladic function which returns data and/or an error. It will
// invoke the function, and return a Value containing either the data or the
// error.
//
// This will panic if `fn` is the wrong type.
func Do(fn interface{}) Value {
	cFn, t, err := callable(fn, 0)
	if err != nil {
		panic(err)
	}

	return retDataToValue(t, cFn.Call(nil))
}

// Data creates a data-Value containing the provided data.
func Data(data ...interface{}) Value {
	ret := make(dataValue, len(data))
	for i, v := range data {
		ret[i] = reflect.ValueOf(v)
	}
	return ret
}

// Error creates an Value containing the provided error.
//
// If the error is nil, this is equivalent to Data() (i.e. a niladic data-Value).
func Error(err error) Value {
	if err == nil {
		return Data()
	}
	return errorValue(reflect.ValueOf(err))
}
