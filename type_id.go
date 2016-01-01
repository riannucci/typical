package typical

import (
	"bytes"
	"reflect"
)

type typeID string

func (t typeID) isErr() bool {
	return t[0] == '!'
}

func dataToTypeID(isErr bool, data []reflect.Value) typeID {
	types := make([]reflect.Type, len(data))
	buf := &bytes.Buffer{}
	if !isErr {
		buf.WriteRune(' ')
	} else {
		buf.WriteRune('!')
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
