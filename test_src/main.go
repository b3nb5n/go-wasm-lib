// go:build js && wasm

package test_src

import "fmt"

func Add(x struct { A, B []int; C *string }, y [8]string) int {
	fmt.Println("")
	return 0
}