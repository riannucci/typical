// Package typical is a weird type-switching library for golang.
//
// It's probably very silly, but I thought it would be a fun project, and
// thought that it could make some really ugly error-handling patterns in go
// much nicer looking.
//
// Performance
//
// Not as bad as I initially suspected! There's a very trivial set of benchmarks
// which are fun to profile and try to optimize. Initially the Typical one ran
// at ~20000ns/op, but now runs at ~8500ns/op on the same hardware. The
// conventional implementation takes ~4000ns/op. Allocation is also not
// horrible; currently `976 B/op   26 allocs/op` vs `504 B/op   10 allocs/op`
// (with go 1.4.2 on linux).
//
// These numbers assume that the methods in the switch statements have been
// preregistered with the RegisterCommonFunction method. Removing this registry
// causes typical to use the slow reflect.Call path.
package typical
