package typical

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"sync"
)

func callable(fn interface{}) (*reflect.Value, reflect.Type) {
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		panic(fmt.Errorf("typical: `fn` has wrong type %T (not a function)", fn))
	}
	t := v.Type()
	return &v, t
}

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
		if t.AssignableTo(inT) {
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

func writeSmallest(w *bytes.Buffer, p uintptr) {
	buf := [8]byte{}
	bs := buf[:]

	switch {
	case p <= math.MaxUint8:
		w.WriteByte(byte(p))
	case p <= math.MaxUint16:
		binary.BigEndian.PutUint16(bs, uint16(p))
		w.Write(bs[:2])
	case p <= math.MaxUint32:
		binary.BigEndian.PutUint32(bs, uint32(p))
		w.Write(bs[:4])
	default:
		binary.BigEndian.PutUint64(bs, uint64(p))
		w.Write(bs)
	}
}

func (v *value) S(funcs ...interface{}) Value {
	for _, fn := range funcs {
		fnV, fnT := callable(fn)
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
	typeOfError = reflect.TypeOf((*error)(nil)).Elem()
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
