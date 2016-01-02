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

func (v *value) S(first interface{}, rest ...interface{}) Value {
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

func retDataToValue(fnTyp reflect.Type, data []reflect.Value) Value {
	if len(data) == 0 {
		return newData(nil)
	}

	lastIdx := fnTyp.NumOut() - 1
	if fnTyp.Out(lastIdx) == typeOfError {
		if !data[lastIdx].IsNil() {
			return newError(data[lastIdx])
		}
		data = data[:lastIdx]
	}
	return newData(data)
}
