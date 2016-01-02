package typical

import (
	"reflect"
	"sync"
)

var mapL = sync.RWMutex{}
var matchMap = map[typeID]map[uintptr]bool{}
var typeMap = map[typeID][]reflect.Type{}

func match(tid typeID, fnT reflect.Type) bool {
	fnID := reflect.ValueOf(fnT).Pointer()

	m, ok, types := func() (m, ok bool, types []reflect.Type) {
		mapL.RLock()
		if m, ok = matchMap[tid][fnID]; !ok {
			types = typeMap[tid]
		}
		mapL.RUnlock()
		return
	}()
	if ok {
		return m
	}

	set := func(m bool) bool {
		mapL.Lock()
		mMap, ok := matchMap[tid]
		if ok {
			mMap[fnID] = m
		} else {
			matchMap[tid] = map[uintptr]bool{fnID: m}
		}
		mapL.Unlock()
		return m
	}

	vt := reflect.Type(nil)
	numIn := fnT.NumIn()
	if fnT.IsVariadic() {
		if numIn-1 > len(types) {
			return set(false)
		}
		vt = fnT.In(numIn - 1).Elem()
		numIn--
	} else if len(types) != numIn {
		return set(false)
	}

	isErr := tid.isErr()
	for i, t := range types {
		inT := reflect.Type(nil)
		if i < numIn {
			inT = fnT.In(i)
		} else {
			if !isErr && vt == typeOfInterface {
				// optimize for ...interface{}
				break
			}
			inT = vt
		}
		if t == inT {
			continue
		}
		if t == nil {
			if inT == typeOfInterface {
				continue
			}
		} else if t.AssignableTo(inT) {
			if !isErr {
				continue
			} else if inT.Implements(typeOfError) {
				continue
			}
		}
		return set(false)
	}

	return set(true)
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
	fnV := reflect.ValueOf(first)
	fnT := fnV.Type()
	if match(v.dataID, fnT) {
		return retDataToValue(fnT, fnV.Call(v.dataErr))
	}

	for _, fn := range rest {
		fnV := reflect.ValueOf(fn)
		fnT := fnV.Type()
		if match(v.dataID, fnT) {
			return retDataToValue(fnT, fnV.Call(v.dataErr))
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

func retDataToValue(fnT reflect.Type, data []reflect.Value) *Value {
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
