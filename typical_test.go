package typical

import (
	"reflect"
	"testing"
)

func BenchmarkTypical(b *testing.B) {
	mapL.Lock()
	matchMap = map[string]map[uintptr]bool{}
	typeMap = map[string][]reflect.Type{}
	mapL.Unlock()

	for i := 0; i < b.N; i++ {

	}
}

func BenchmarkConventional(b *testing.B) {
}
