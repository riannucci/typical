package typical

import (
	"fmt"
	"os"
	"reflect"
)

// EnableNotify makes typical print a line to stdout for every function it
// calls which is not registered in commonFunctions.
var EnableNotify = false

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
func (v Value) S(first interface{}, rest ...interface{}) Value {
	matchFn := match
	if v.isErr {
		matchFn = matchErr
	}

	fnV := reflect.ValueOf(first)
	fnT := fnV.Type()
	if matchFn(fnT, v.dataErr) {
		return v.call(fnV, fnT)
	}

	for i := range rest {
		fnV = reflect.ValueOf(rest[i])
		fnT = fnV.Type()
		if matchFn(fnT, v.dataErr) {
			return v.call(fnV, fnT)
		}
	}

	return v
}

var (
	typeOfError = reflect.TypeOf((*error)(nil)).Elem()
)

func (v Value) call(fnV reflect.Value, fnT reflect.Type) Value {
	data := []reflect.Value(nil)
	if cmn, ok := commonFunctions[fnT]; ok {
		data = cmn(fnV.Interface(), v.dataErr)
	} else {
		if EnableNotify {
			fmt.Fprintf(os.Stderr, "typical: function not registered: %s\n", fnT)
		}
		data = fnV.Call(v.dataErr)
	}

	if len(data) == 0 {
		return newData(nil)
	}

	lastIdx := fnT.NumOut() - 1
	if fnT.Out(lastIdx) == typeOfError {
		if !data[lastIdx].IsNil() {
			return newError(data[lastIdx])
		}
		data = data[:lastIdx]
	}
	return newData(data)
}
