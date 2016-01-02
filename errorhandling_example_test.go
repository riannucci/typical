package typical

import (
	"fmt"
	"reflect"
)

type mySpecialError struct{ msg string }

func (m *mySpecialError) Error() string {
	return fmt.Sprintf("special: %s", m.msg)
}

func handle(fn interface{}) {
	Do(fn).S(
		func(i int) {
			fmt.Println("got int", i)
		},
		func(v interface{}) {
			// notice how this doesn't catch the errors
			fmt.Printf("cannot handle data: (%T)(%+v)\n", v, v)
		},
		func() {
			fmt.Println("got nothin")
		},

		// error handling functions
		func(s *mySpecialError) {
			fmt.Println("ignoring:", s)
		},
		func(err error) {
			fmt.Println("got other error:", err)
		},
	)
}

func init() {
	RegisterCommonFunction((func(interface{}))(nil), func(fnI interface{}, in []reflect.Value) []reflect.Value {
		f := fnI.(func(interface{}))
		f(in[0].Interface())
		return nil
	})
	RegisterCommonFunction((func(int))(nil), func(fnI interface{}, in []reflect.Value) []reflect.Value {
		f := fnI.(func(int))
		f(int(in[0].Int()))
		return nil
	})
}

// Shows how error handling can work
func ExampleDo_errorHandling() {
	handle(func() int { return 10 })
	handle(func() error { return fmt.Errorf("generic error") })
	handle(func() error { return nil })
	handle(func() error { return &mySpecialError{"waffle"} })
	handle(func() string { return "Some Name" })

	// Output:
	// got int 10
	// got other error: generic error
	// got nothin
	// ignoring: special: waffle
	// cannot handle data: (string)(Some Name)
}
