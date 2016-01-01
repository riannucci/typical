package typical

import (
	"fmt"
	"reflect"
)

func callable(fn interface{}, numIn int) (*reflect.Value, reflect.Type, error) {
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		panic(fmt.Errorf("typical: `fn` has wrong type %T (not a function)", fn))
	}
	t := v.Type()
	if t.NumIn() != numIn {
		return nil, nil, fmt.Errorf("typical: `fn` has wrong type %T (too many inputs)", fn)
	}
	return &v, t, nil
}

func doSwitch(data []reflect.Value, funcs []interface{}) Value {
	numIn := len(data)
	callData := make([]reflect.Value, len(data))
	inTyps := []reflect.Type(nil)
	getInTyps := func() (typs []reflect.Type) {
		if inTyps == nil {
			inTyps = make([]reflect.Type, len(data))
			for i, d := range data {
				if d.Kind() == reflect.Interface {
					if !d.IsNil() {
						inTyps[i] = d.Elem().Type()
					}
					// otherwise the TYPE is nil
				} else {
					inTyps[i] = d.Type()
				}
			}
		}
		return inTyps
	}

	for _, fn := range funcs {
		cFn, ct, err := callable(fn, numIn)
		if err != nil {
			continue
		}
		inTyps := getInTyps()
		match := true
		for i := range inTyps {
			candidate := ct.In(i)
			concrete := inTyps[i]
			if candidate == concrete {
				callData[i] = data[i]
				continue
			}
			if candidate == typeOfNilValue && data[i].IsNil() {
				callData[i] = valueOfNilValue
				continue
			}
			if candidate.Kind() == reflect.Interface && concrete != nil && concrete.Implements(candidate) {
				callData[i] = data[i].Convert(candidate)
				continue
			}
			match = false
			break
		}

		if match {
			return retDataToValue(ct, cFn.Call(callData))
		}
	}

	return nil
}

func notNillableOrNotNil(v *reflect.Value) bool {
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Slice:
		return !v.IsNil()
	}
	return true
}

func retDataToValue(fnTyp reflect.Type, data []reflect.Value) Value {
	lastIdx := fnTyp.NumOut() - 1
	if fnTyp.Out(lastIdx) == typeOfError {
		if lastDatum := data[lastIdx]; notNillableOrNotNil(&lastDatum) {
			return errorValue(lastDatum)
		}
		return dataValue(data[:lastIdx])
	}
	return dataValue(data)
}
