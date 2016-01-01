package typical

import (
	"fmt"
	"strings"
)

func sum(vals ...interface{}) (interface{}, error) {
	return Data(vals...).S(
		func(a, b int) int {
			return a + b
		},
		func(a, b float64) float64 {
			return a + b
		},
		func(a string, b error) string {
			return a + b.Error()
		},
		func(a string) string {
			return a
		},
		func(a string, rest ...string) string {
			return strings.ToUpper(a) + " " + strings.Join(rest, " ")
		},
		func(a, b interface{}) error {
			return fmt.Errorf("unsupported types %T and %T", a, b)
		},
		func(vals ...interface{}) error {
			return fmt.Errorf("unsupported types:", vals...)
		}).FirstErr()
}

// Shows the power of pattern matching
func ExampleDo_patternMatch() {
	fmt.Println(sum(1, 2))
	fmt.Println(sum(1.0, 2.5))
	fmt.Println(sum("a", "b", "c"))
	fmt.Println(sum("hi", fmt.Errorf(" there")))

	fmt.Println(sum(1.0, 2))
	fmt.Println(sum("cat", 2.0))
	fmt.Println(sum("cat", 2.0, 10))
	fmt.Println(sum(nil))
	fmt.Println(sum((*int)(nil)))
	fmt.Println(sum())

	// Output:
	// 3 <nil>
	// 3.5 <nil>
	// A b c <nil>
	// hi there <nil>
	// <nil> unsupported types float64 and int
	// <nil> unsupported types string and float64
	// <nil> unsupported types:%!(EXTRA string=cat, float64=2, int=10)
	// <nil> unsupported types:%!(EXTRA <nil>)
	// <nil> unsupported types:%!(EXTRA *int=<nil>)
	// <nil> unsupported types:
}
