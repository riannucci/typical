package typical

import (
	"reflect"
	"testing"
)

func BenchmarkTypical(b *testing.B) {
	mapL.Lock()
	matchMap = map[typeID]map[uintptr]bool{}
	typeMap = map[typeID][]reflect.Type{}
	mapL.Unlock()

	for i := 0; i < b.N; i++ {

	}
}

func BenchmarkConventional(b *testing.B) {
}
