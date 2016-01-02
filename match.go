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
		numIn--
		if numIn > len(data) {
			return false
		}
		if numIn < len(data) {
			vt = fnT.In(numIn).Elem()
		}
	} else if len(data) != numIn {
		return false
	}

	i := 0
	t := reflect.Type(nil)
	inT := vt
	for ; i < numIn; i++ {
		t = data[i].Type()
		inT = fnT.In(i)
		if (t == nil && inT == typeOfInterface) || (t != nil && t.AssignableTo(inT)) {
			continue
		}
		return false
	}
	if vt != nil {
		if vt == typeOfInterface {
			// quick optimization for ...interface{}
			return true
		}
		for ; i < len(data); i++ {
			t = data[i].Type()
			if t != nil && t.AssignableTo(vt) {
				continue
			}
			return false
		}
	}

	return true
}
