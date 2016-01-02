package typical

import (
	"reflect"
)

var (
	typeOfInterface = reflect.TypeOf((*interface{})(nil)).Elem()
)

func matchErr(fnT reflect.Type, data []reflect.Value) bool {
	if fnT.NumIn() != 1 {
		return false
	}
	inT := fnT.In(0)
	if !inT.Implements(typeOfError) {
		return false
	}
	return data[0].Type().AssignableTo(inT)
}

func match(fnT reflect.Type, data []reflect.Value) bool {
	vt := reflect.Type(nil)
	numIn := fnT.NumIn()
	if fnT.IsVariadic() {
		if numIn-1 > len(data) {
			return false
		}
		vt = fnT.In(numIn - 1).Elem()
		numIn--
	} else if len(data) != numIn {
		return false
	}

	t := reflect.Type(nil)
	inT := vt
	for i := range data {
		t = data[i].Type()
		inT = vt
		if i < numIn {
			inT = fnT.In(i)
		} else if vt == typeOfInterface {
			// optimize for ...interface{}
			return true
		}
		if t == inT || (t == nil && inT == typeOfInterface) || (t != nil && t.AssignableTo(inT)) {
			continue
		}
		return false
	}
	return true
}
