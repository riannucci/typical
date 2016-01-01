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
var matchMap = map[string]map[uintptr]bool{}
var typeMap = map[string][]reflect.Type{}

func match(typeID string, fnT reflect.Type) bool {
	fnID := reflect.ValueOf(fnT).Pointer()

	m, ok, types := func() (m, ok bool, types []reflect.Type) {
		mapL.RLock()
		if m, ok = matchMap[typeID][fnID]; !ok {
			types = typeMap[typeID]
		}
		mapL.RUnlock()
		return
	}()
	if ok {
		return m
	}

	set := func(m bool) bool {
		mapL.Lock()
		mMap, ok := matchMap[typeID]
		if ok {
			mMap[fnID] = m
		} else {
			matchMap[typeID] = map[uintptr]bool{fnID: m}
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

	for i, t := range types {
		inT := reflect.Type(nil)
		if i < numIn {
			inT = fnT.In(i)
		} else {
			inT = vt
		}
		if t == inT || t.AssignableTo(inT) {
			continue
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

func dataToTypeID(data []reflect.Value) string {
	types := make([]reflect.Type, len(data))
	buf := &bytes.Buffer{}
	for i, d := range data {
		if d.IsValid() {
			t := d.Type()
			types[i] = t
			writeSmallest(buf, reflect.ValueOf(t).Pointer())
		} else {
			writeSmallest(buf, 0)
		}
	}
	ret := buf.String()
	mapL.Lock()
	typeMap[ret] = types
	mapL.Unlock()
	return ret
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
	lastIdx := fnTyp.NumOut() - 1
	if fnTyp.Out(lastIdx) == typeOfError {
		if notNillableOrNotNil(&data[lastIdx]) {
			return newValue(true, data[lastIdx:])
		}
		return newValue(false, data[:lastIdx])
	}
	return newValue(false, data)
}
