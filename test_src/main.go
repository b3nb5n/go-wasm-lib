// go:build js && wasm

package test_src

func Add(x struct { A, B []int; C *string }, y [8]string) int {
	return 0
}