package typical

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"sync"
)

var mapL = sync.RWMutex{}

type matchKey struct {
	fromTypes typeID
	toFn      uintptr
}

var matchMap = map[matchKey]bool{}
var typeMap = map[typeID][]reflect.Type{}

type typeID interface {
	iAmTypeID()

	isErr() bool
}

type staticFnTypeID uintptr

func (s staticFnTypeID) iAmTypeID()  {}
func (s staticFnTypeID) isErr() bool { return false }

type errorTypeID uintptr

func (s errorTypeID) iAmTypeID()  {}
func (s errorTypeID) isErr() bool { return true }

type dataTypeID string

func (s dataTypeID) iAmTypeID()  {}
func (s dataTypeID) isErr() bool { return false }

func writeSmallest(w *bytes.Buffer, p uintptr) {
	buf := [8]byte{}
	binary.BigEndian.PutUint64(buf[:], uint64(p))
	i := 0
	for ; i < len(buf)-1; i++ {
		if buf[i] != 0 {
			break
		}
	}
	w.Write(buf[i:])
}

func addTypeMap(id typeID, types []reflect.Type) {
	mapL.Lock()
	if _, ok := typeMap[id]; !ok {
		typeMap[id] = types
	}
	mapL.Unlock()
}

func getMatchData(key matchKey) (match, ok bool, types []reflect.Type) {
	mapL.RLock()
	match, ok = matchMap[key]
	if !ok {
		types = typeMap[key.fromTypes]
	}
	mapL.RUnlock()
	return
}

func setMatchMap(key matchKey, val bool) bool {
	mapL.Lock()
	if _, ok := matchMap[key]; !ok {
		matchMap[key] = val
	}
	mapL.Unlock()
	return val
}

func errToTypeID(err []reflect.Value) typeID {
	t := err[0].Type()
	types := []reflect.Type{t}
	ret := errorTypeID(reflect.ValueOf(t).Pointer())
	addTypeMap(ret, types)
	return ret
}

func dataToTypeID(fnT reflect.Type, data []reflect.Value) typeID {
	ret := typeID(nil)

	concrete := fnT != nil
	if concrete {
		ret = staticFnTypeID(reflect.ValueOf(fnT).Pointer())
		mapL.RLock()
		_, ok := typeMap[ret]
		mapL.RUnlock()
		if ok {
			return ret
		}
	}

	buf := &bytes.Buffer{}
	types := []reflect.Type(nil)
	if amt := len(data); amt > 0 {
		types = make([]reflect.Type, amt)
		for i, d := range data {
			if d.Kind() == reflect.Interface {
				concrete = false
				if d = d.Elem(); d.IsValid() {
					data[i] = d
				}
			}
			if d.IsValid() {
				t := d.Type()
				types[i] = t
				writeSmallest(buf, reflect.ValueOf(t).Pointer())
			} else {
				writeSmallest(buf, 0)
			}
		}
	}

	if !concrete {
		ret = dataTypeID(buf.String())
	}

	addTypeMap(ret, types)
	return ret
}
