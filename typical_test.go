package typical

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// ExampleDo_json shows how to use typical to build an IO routine nicely.
func ExampleDo_json() {
	type SomeObject struct {
		Field string `json:",omitempty"`
	}

	// normally you'd use a json.Encoder to write directly to the stream, but
	// by way of example, we serialize to bytes separately first.
	writeJsonToStream := func(obj interface{}, w io.Writer) error {
		return Data(obj).S(json.Marshal).S(func(data []byte) ([]byte, error) {
			if len(data) < 10 {
				// impose silly error condition that data must be > 10 bytes
				return nil, fmt.Errorf("data too short!")
			}
			return data, nil
		}).S(w.Write).S(func(_ int) error { // ignore the int data
			_, err := w.Write([]byte("\n"))
			return err
		}).Error()
	}

	normalImpl := func(obj interface{}, w io.Writer) error {
		data, err := json.Marshal(obj)
		if err != nil {
			return err
		}
		if len(data) < 10 {
			return fmt.Errorf("data too short!")
		}
		_, err = w.Write(data)
		if err != nil {
			return err
		}
		_, err = w.Write([]byte("\n"))
		return err
	}
	_ = normalImpl

	fmt.Println("error:", writeJsonToStream(&SomeObject{}, os.Stdout))
	fmt.Println("error:", writeJsonToStream(complex(1, 2), os.Stdout))
	fmt.Println("error:", writeJsonToStream(&SomeObject{"hello world!"}, os.Stdout))
	// Output:
	// error: data too short!
	// error: json: unsupported type: complex128
	// {"Field":"hello world!"}
	// error: <nil>
}
