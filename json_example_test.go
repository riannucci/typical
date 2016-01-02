package typical

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
)

// normally you'd use a json.Encoder to write directly to the stream, but
// by way of example, we serialize to bytes separately first.
func writeJsonToStream(obj interface{}, w io.Writer) error {
	return Data(obj).S(
		json.Marshal,
	).S(
		w.Write,
	).S(func(_ int) error { // ignore the int data
		_, err := w.Write([]byte("\n"))
		return err
	}).Error()
}

func init() {
	RegisterCommonFunction((func(int) error)(nil), func(fnI interface{}, in []reflect.Value) []reflect.Value {
		f := fnI.(func(int) error)
		return IfaceToValues(f(int(in[0].Int())))
	})
}

func normalJsonWriteFunction(obj interface{}, w io.Writer) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	if _, err = w.Write(data); err != nil {
		return err
	}
	_, err = w.Write([]byte("\n"))
	return err
}

type someObject struct {
	Field string `json:",omitempty"`
}

// Shows how to use typical to build an IO routine nicely.
func ExampleDo_json() {
	fmt.Println("error:", writeJsonToStream(complex(1, 2), os.Stdout))
	fmt.Println("error:", writeJsonToStream(&someObject{"hello world!"}, os.Stdout))
	fmt.Println("error:", writeJsonToStream(&someObject{"hello world again!"}, os.Stdout))
	// Output:
	// error: json: unsupported type: complex128
	// {"Field":"hello world!"}
	// error: <nil>
	// {"Field":"hello world again!"}
	// error: <nil>
}
