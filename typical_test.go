package typical

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Shows how to use typical to build an IO routine nicely.
func ExampleDo_json() {
	type SomeObject struct {
		Field string `json:",omitempty"`
	}

	// normally you'd use a json.Encoder to write directly to the stream, but
	// by way of example, we serialize to bytes separately first.
	writeJsonToStream := func(obj interface{}, w io.Writer) error {
		return Data(obj).S(
			json.Marshal,
		).S(
			w.Write,
		).S(func(_ int) error { // ignore the int data
			_, err := w.Write([]byte("\n"))
			return err
		}).Error()
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
	// Output:
	// error: json: unsupported type: complex128
	// {"Field":"hello world!"}
	// error: <nil>
}

// Shows the power of pattern matching
func ExampleDo_patternMatch() {
	sum := func(a, b interface{}) (interface{}, error) {
		return Data(a, b).S(
			func(a, b int) int {
				return a + b
			},
			func(a, b float64) float64 {
				return a + b
			},
			func(a, b interface{}) error {
				return fmt.Errorf("unsupported types %T and %T", a, b)
			}).FirstErr()
	}

	fmt.Println(sum(1, 2))
	fmt.Println(sum(1.0, 2.5))

	fmt.Println(sum(1.0, 2))
	fmt.Println(sum("cat", 2.0))

	// Output:
	// 3 <nil>
	// 3.5 <nil>
	// <nil> unsupported types float64 and int
	// <nil> unsupported types string and float64
}
