package typical

import (
	"bytes"
	"encoding/binary"
	"reflect"
)

type typeID string

func (t typeID) isErr() bool {
	return t[0] == 0xff
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

func dataToTypeID(isErr bool, data []reflect.Value) typeID {
	types := make([]reflect.Type, len(data))
	buf := &bytes.Buffer{}
	if !isErr {
		buf.WriteByte(0)
	} else {
		buf.WriteByte(0xff)
	}
	for i, d := range data {
		if d.Kind() == reflect.Interface {
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
	ret := typeID(buf.String())
	mapL.Lock()
	typeMap[ret] = types
	mapL.Unlock()
	return ret
}
