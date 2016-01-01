package typical

import (
	"reflect"
)

type value struct {
	// contains the multiple data or the singular error
	dataErr []reflect.Value
	dataID  typeID
}

var _ Value = (*value)(nil)

func (v *value) First() interface{} {
	r, err := v.FirstErr()
	if err != nil {
		panic(err)
	}
	return r
}

func (v *value) FirstErr() (interface{}, error) {
	if err := v.Error(); err != nil {
		return nil, err
	}
	first := v.dataErr[0]
	if notNillableOrNotNil(&first) {
		return first.Interface(), nil
	}
	return nil, nil
}

func (v *value) All() []interface{} {
	r, err := v.AllErr()
	if err != nil {
		panic(err)
	}
	return r
}

func (v *value) AllErr() ([]interface{}, error) {
	if err := v.Error(); err != nil {
		return nil, err
	}
	ret := make([]interface{}, len(v.dataErr))
	for i, v := range v.dataErr {
		if notNillableOrNotNil(&v) {
			ret[i] = v.Interface()
		}
	}
	return ret, nil
}

func (v *value) Error() error {
	if v.dataID.isErr() {
		return v.dataErr[0].Interface().(error)
	}
	return nil
}

func newData(data []reflect.Value) *value {
	return &value{data, dataToTypeID(false, data)}
}

func newError(err reflect.Value) *value {
	data := []reflect.Value{err}
	return &value{data, dataToTypeID(true, data)}
}
