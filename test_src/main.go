// go:build js && wasm

package test_src

func Add(a, b int64, _ float64) int64 {
	return a + b
}

func Hello(name string) {}