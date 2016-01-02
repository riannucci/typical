package typical

import (
	"bytes"
	"encoding/binary"
	"reflect"
)

type typeIDIndicator byte

const (
	typeIDData       = 0x00
	typeIDStaticData = 0x55
	typeIDError      = 0xff
)

type typeID string

func (t typeID) isErr() bool {
	return t[0] == typeIDError
}

func writeSmallest(w *bytes.Buffer, p uintptr) {
	buf := [8]byte{}
	binary.BigEndian.PutUint64(buf[:], uint64(p))
	i := 0
	for ; i < len(buf); i++ {
		if buf[i] != 0 {
			break
		}
	}
	w.Write(buf[i:])
}

func dataToTypeID(isErr bool, fnT reflect.Type, data []reflect.Value) typeID {
	buf := &bytes.Buffer{}

	ret := typeID("")

	concrete := fnT != nil
	if concrete {
		buf.WriteByte(typeIDStaticData)
		writeSmallest(buf, reflect.ValueOf(fnT).Pointer())
		ret = typeID(buf.String())
		mapL.RLock()
		_, ok := typeMap[ret]
		mapL.RUnlock()
		if ok {
			return ret
		}
		buf.Reset()
	}

	if !isErr {
		buf.WriteByte(typeIDData)
	} else {
		buf.WriteByte(typeIDError)
	}

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
		ret = typeID(buf.String())
	}

	mapL.Lock()
	if _, ok := typeMap[ret]; !ok {
		typeMap[ret] = types
	}
	mapL.Unlock()
	return ret
}
