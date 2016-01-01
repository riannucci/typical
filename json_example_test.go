package typical

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
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

// Shows how to use typical to build an IO routine nicely.
func ExampleDo_json() {
	type SomeObject struct {
		Field string `json:",omitempty"`
	}

	normalImpl := func(obj interface{}, w io.Writer) error {
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
	_ = normalImpl

	fmt.Println("error:", writeJsonToStream(complex(1, 2), os.Stdout))
	fmt.Println("error:", writeJsonToStream(&SomeObject{"hello world!"}, os.Stdout))
	fmt.Println("error:", writeJsonToStream(&SomeObject{"hello world again!"}, os.Stdout))
	// Output:
	// error: json: unsupported type: complex128
	// {"Field":"hello world!"}
	// error: <nil>
	// {"Field":"hello world again!"}
	// error: <nil>
}
