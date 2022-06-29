// go:build js && wasm

package test_src

func Hello(name string) string {
	return "hello " + name
}